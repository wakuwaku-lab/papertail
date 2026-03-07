package cmd

import (
	"github.com/spf13/cobra"
)

var cfgFile string

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.papertail.yaml)")
}

var rootCmd = &cobra.Command{
	Use:   "papertail",
	Short: "Papertail - Download and process Papertrail logs",
	Long: `Papertail is a CLI tool for downloading and processing logs from Papertrail API.

It supports downloading logs by hour or date range and optionally importing them into SQLite.

Required environment variable:
  PAPERTRAIL_API_TOKEN - Your Papertrail API token

Example:
  export PAPERTRAIL_API_TOKEN=your_token_here
  papertail logs --db logs.db --tsv output.tsv --start 1 --end 24`,
}
