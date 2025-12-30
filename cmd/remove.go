// cmd/remove.go
package cmd

import (
	"fmt"

	"github.com/jpequegn/xmon/internal/account"
	"github.com/jpequegn/xmon/internal/config"
	"github.com/jpequegn/xmon/internal/database"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <username>",
	Short: "Remove an X account from monitoring",
	Long:  `Removes an X account from your monitoring list.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

func runRemove(cmd *cobra.Command, args []string) error {
	username := args[0]

	db, err := database.New(config.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	repo := account.NewRepository(db)
	if !repo.Exists(username) {
		return fmt.Errorf("account @%s is not being monitored", username)
	}

	if err := repo.Remove(username); err != nil {
		return fmt.Errorf("failed to remove account: %w", err)
	}

	fmt.Printf("Removed @%s from monitoring\n", username)
	return nil
}
