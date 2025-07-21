package cmd

import (
	"fmt"
	"os"
	"sort"
	"time"

	"ademun/md5/utils"

	"github.com/spf13/cobra"
)

var (
	profilingMode bool
)

var rootCmd = &cobra.Command{
	Use:   "md5",
	Short: "Returns the md5 hash of files / file in the specified directory / file",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Please specify a path to directory / file")
			return
		}
		var profiler utils.Profiler
		if profilingMode {
			profiler = *utils.NewProfiler()
			profiler.Start()
			defer func() {
				profiler.Stop()
				profiler.Stat(time.Second)
			}()
		}
		data, err := Parse(args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		files := make([]string, 0)
		for k := range data {
			files = append(files, k)
		}
		sort.Strings(files)
		for _, k := range files {
			fmt.Printf("%s\t%s\n", data[k], k)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolVarP(&profilingMode, "profiling", "p", false, "measures total working time and average goroutine number")
}
