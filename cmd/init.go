// cmd/init.go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jpequegn/xmon/internal/config"
	"github.com/jpequegn/xmon/internal/database"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize xmon configuration",
	Long:  `Creates the config directory, database, and prompts for your X API bearer token.`,
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Initializing xmon...")

	// Create config directory
	if err := os.MkdirAll(config.ConfigDir(), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	fmt.Printf("Created config directory: %s\n", config.ConfigDir())

	// Initialize database
	db, err := database.New(config.DBPath())
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	db.Close()
	fmt.Printf("Created database: %s\n", config.DBPath())

	// Load or create config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Prompt for bearer token if not set
	if cfg.X.BearerToken == "" {
		fmt.Print("\nEnter your X API Bearer Token: ")
		reader := bufio.NewReader(os.Stdin)
		token, _ := reader.ReadString('\n')
		token = strings.TrimSpace(token)

		if token != "" {
			cfg.X.BearerToken = token
		}
	}

	// Save config
	if err := cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Saved config: %s\n", config.ConfigPath())

	fmt.Println("\nxmon initialized successfully!")
	fmt.Println("Next: Run 'xmon add <username>' to add accounts to monitor.")

	return nil
}
