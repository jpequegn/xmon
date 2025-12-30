// cmd/show.go
package cmd

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jpequegn/xmon/internal/account"
	"github.com/jpequegn/xmon/internal/config"
	"github.com/jpequegn/xmon/internal/database"
	"github.com/jpequegn/xmon/internal/tweet"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <username>",
	Short: "Show details for an account",
	Long:  `Displays detailed activity for a specific monitored account.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

var (
	showDays int
)

func init() {
	rootCmd.AddCommand(showCmd)
	showCmd.Flags().IntVar(&showDays, "days", 7, "Number of days to show")
}

func runShow(cmd *cobra.Command, args []string) error {
	username := args[0]

	db, err := database.New(config.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	accountRepo := account.NewRepository(db)
	tweetRepo := tweet.NewRepository(db)

	acc, err := accountRepo.Get(username)
	if err != nil {
		return fmt.Errorf("account @%s not found", username)
	}

	since := time.Now().AddDate(0, 0, -showDays)

	tweets, _ := tweetRepo.GetForAccount(acc.ID, since)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fmt.Printf("\n%s\n", titleStyle.Render("@"+acc.Username))
	if acc.Name != "" {
		fmt.Println(acc.Name)
	}
	if acc.Bio != "" {
		fmt.Printf("%s\n", dimStyle.Render(acc.Bio))
	}
	fmt.Printf("%s\n\n", dimStyle.Render(fmt.Sprintf("%d followers", acc.Followers)))

	// Count by type
	originals, retweets, quotes := 0, 0, 0
	for _, t := range tweets {
		switch t.TweetType {
		case "original":
			originals++
		case "retweet":
			retweets++
		case "quote":
			quotes++
		}
	}

	fmt.Printf("%s (last %d days)\n", sectionStyle.Render("Activity"), showDays)
	fmt.Printf("  Originals: %d\n", originals)
	fmt.Printf("  Retweets:  %d\n", retweets)
	fmt.Printf("  Quotes:    %d\n", quotes)
	fmt.Println()

	// Recent tweets
	if len(tweets) > 0 {
		fmt.Printf("%s\n", sectionStyle.Render("Recent Tweets"))
		limit := 5
		if len(tweets) < limit {
			limit = len(tweets)
		}
		for i := 0; i < limit; i++ {
			t := tweets[i]
			content := t.Content
			if len(content) > 70 {
				content = content[:67] + "..."
			}
			fmt.Printf("  %s %s\n", dimStyle.Render(t.CreatedAt.Format("Jan 2")), content)
		}
	}

	return nil
}
