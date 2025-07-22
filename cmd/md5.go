package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

var numDigesters = 5
var chunkSize = 1024

func Parse(root string) (map[string]string, error) {
	m := make(map[string]string)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	paths, errc := walkRoot(ctx, root)

	res := digest(ctx, root, paths, numDigesters)

	for r := range res {
		if r.err != nil && !errors.Is(r.err, fs.ErrPermission) {
			m[r.path] = r.err.Error()
		}
		m[r.path] = hex.EncodeToString(r.sum[:])
	}

	select {
	case err := <-errc:
		if err != nil {
			return nil, err
		}
	default:
	}
	return m, nil
}

// Ignores nested dirs and symlinks
func walkRoot(ctx context.Context, root string) (<-chan string, chan error) {
	paths := make(chan string)
	errc := make(chan error, 1)
	go func() {
		defer func() {
			close(paths)
			close(errc)
		}()
		fileInfo, err := os.Stat(root)
		if err != nil {
			errc <- err
			return
		}
		if fileInfo.Mode().IsRegular() {
			paths <- root
			return
		}
		errc <- filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path == root {
				return nil
			}
			if !d.Type().IsRegular() || d.Type()&os.ModeSymlink != 0 {
				return fs.SkipDir
			}
			select {
			case <-ctx.Done():
				return errors.New("walk cancelled")
			case paths <- path:
			}
			return nil
		})
	}()
	return paths, errc
}

func digest(ctx context.Context, root string, paths <-chan string, numDigesters int) <-chan result {
	res := make(chan result)
	wg := &sync.WaitGroup{}
	wg.Add(numDigesters)
	go func() {
		for range numDigesters {
			go func() {
				defer wg.Done()
				digester(ctx, root, paths, res)
			}()
		}
		wg.Wait()
		close(res)
	}()
	return res
}

func digester(ctx context.Context, root string, paths <-chan string, res chan<- result) {
	for p := range paths {
		hashSum, hashErr := md5Sum(p)
		rel, pathErr := filepath.Rel(root, p)
		if root == p || pathErr != nil {
			rel = root
		}
		select {
		case <-ctx.Done():
			return
		default:
			if hashErr != nil {
				res <- result{err: hashErr}
				continue
			}
			res <- result{path: rel, sum: hashSum, err: pathErr}
		}
	}
}

// Process file in chunks
func md5Sum(path string) ([md5.Size]byte, error) {
	buf := make([]byte, chunkSize)
	hash := make([]byte, 0)
	file, err := os.Open(path)
	if err != nil {
		return [md5.Size]byte{}, err
	}
	for {
		_, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				file.Close()
				break
			}
			return [md5.Size]byte{}, err
		}
		sum := md5.Sum(buf)
		hash = hex.AppendEncode(hash, sum[:])
	}
	hashSum := md5.Sum(hash)
	return hashSum, nil
}
