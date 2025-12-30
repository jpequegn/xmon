// cmd/add.go
package cmd

import (
	"fmt"

	"github.com/jpequegn/xmon/internal/account"
	"github.com/jpequegn/xmon/internal/config"
	"github.com/jpequegn/xmon/internal/database"
	"github.com/jpequegn/xmon/internal/x"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <username>",
	Short: "Add an X account to monitor",
	Long:  `Adds an X account to your monitoring list by username (without @).`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	username := args[0]

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w (run 'xmon init' first)", err)
	}

	if cfg.X.BearerToken == "" {
		return fmt.Errorf("X API bearer token not set. Add it to %s", config.ConfigPath())
	}

	db, err := database.New(config.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	repo := account.NewRepository(db)
	if repo.Exists(username) {
		return fmt.Errorf("account @%s is already being monitored", username)
	}

	// Fetch user info from X API
	client := x.NewClient(cfg.X.BearerToken)
	user, err := client.GetUser(username)
	if err != nil {
		return fmt.Errorf("failed to fetch user @%s: %w", username, err)
	}

	// Add to database
	if err := repo.Add(user.ID, user.Username, user.Name, user.Description, user.PublicMetrics.FollowersCount); err != nil {
		return fmt.Errorf("failed to add account: %w", err)
	}

	fmt.Printf("Added @%s (%s) - %d followers\n", user.Username, user.Name, user.PublicMetrics.FollowersCount)
	return nil
}
