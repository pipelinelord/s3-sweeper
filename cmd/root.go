/*
Copyright © 2025 CHAMATH PRAMODAYA <chamathrko@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "s3-sweeper",
	Short: "A high-performance S3 stale object analyzer",
	Long: `s3-sweeper is a CLI tool designed to demonstrate Go concurrency patterns.
It scans S3 buckets using a worker pool to identify stale resources efficiently.`,
	
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global Flags
	// We allow the user to specify the AWS Region globally.
	rootCmd.PersistentFlags().String("region", "us-east-1", "AWS Region (default is us-east-1)")
}
