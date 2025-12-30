// cmd/accounts.go
package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/jpequegn/xmon/internal/account"
	"github.com/jpequegn/xmon/internal/config"
	"github.com/jpequegn/xmon/internal/database"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:   "accounts",
	Short: "List monitored accounts",
	Long:  `Shows all X accounts you are currently monitoring.`,
	RunE:  runAccounts,
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}

func runAccounts(cmd *cobra.Command, args []string) error {
	db, err := database.New(config.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	repo := account.NewRepository(db)
	accounts, err := repo.List()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts being monitored.")
		fmt.Println("Run 'xmon add <username>' to add accounts.")
		return nil
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fmt.Printf("\n%s\n\n", titleStyle.Render("Monitored Accounts"))

	for _, acc := range accounts {
		fmt.Printf("  %s", userStyle.Render("@"+acc.Username))
		if acc.Name != "" {
			fmt.Printf(" (%s)", acc.Name)
		}
		fmt.Println()
		fmt.Printf("    %s\n", dimStyle.Render(fmt.Sprintf("%d followers", acc.Followers)))
	}

	fmt.Printf("\n%s\n", dimStyle.Render(fmt.Sprintf("Total: %d accounts", len(accounts))))

	return nil
}
