# MD5 CLI Tool

A concise Go-based command-line utility for computing MD5 file hashes.

## Features
- Parallel file processing
- Directory and single file support

## Usage
```bash
md5 /path/to/directory
md5 /path/to/file.txt
md5 -p /path  # with profiling
```

## Dependencies
- [cobra](https://github.com/spf13/cobra) - CLI framework