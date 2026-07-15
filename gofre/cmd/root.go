package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gofre",
	Short: "Build Python packages with Go extensions",
	Long: `GoFre is a build system for creating Python packages that bundle
Go binaries and native extension modules using pure Go (no CGo required).
Inspired by maturin for Rust/Python.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
}
