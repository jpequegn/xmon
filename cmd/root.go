// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "xmon",
	Short: "Monitor X/Twitter accounts and get activity digests",
	Long:  `xmon helps you track influential X accounts, digest their tweets, and surface early signals about emerging topics.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
