package main

import (
	"fmt"
	"os"

	"github.com/lonegunmanb/hclmerge/pkg"
	"github.com/spf13/cobra"
)

func main() {
	var file1, file2, destFile string

	var rootCmd = &cobra.Command{
		Use:   "hclmerge",
		Short: "Merge two HCL config files",
		Long:  `Merge two HCL config files using hclwrite, save merged content into new file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pkg.MergeFile(file1, file2, destFile)
		},
	}

	rootCmd.Flags().StringVarP(&file1, "file1", "1", "", "config file 1 (required)")
	rootCmd.Flags().StringVarP(&file2, "file2", "2", "", "config file 2 (required)")
	rootCmd.Flags().StringVarP(&destFile, "dest", "d", "", "file to save merged content (required)")
	_ = rootCmd.MarkFlagRequired("file1")
	_ = rootCmd.MarkFlagRequired("file2")
	_ = rootCmd.MarkFlagRequired("dest")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
