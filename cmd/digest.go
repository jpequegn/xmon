// cmd/digest.go
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

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Show activity digest",
	Long:  `Displays a summary of recent tweets from monitored accounts.`,
	RunE:  runDigest,
}

var (
	digestDays int
)

func init() {
	rootCmd.AddCommand(digestCmd)
	digestCmd.Flags().IntVar(&digestDays, "days", 7, "Number of days to include in digest")
}

func runDigest(cmd *cobra.Command, args []string) error {
	db, err := database.New(config.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	since := time.Now().AddDate(0, 0, -digestDays)
	endDate := time.Now()

	accountRepo := account.NewRepository(db)
	tweetRepo := tweet.NewRepository(db)

	accounts, _ := accountRepo.List()
	accountMap := make(map[int64]*account.Account)
	for i := range accounts {
		accountMap[accounts[i].ID] = &accounts[i]
	}

	originals, retweets, quotes, _ := tweetRepo.CountByType(since)
	totalTweets := originals + retweets + quotes

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	fmt.Printf("\n%s (%s - %s)\n",
		titleStyle.Render("X DIGEST"),
		since.Format("Jan 2"),
		endDate.Format("Jan 2, 2006"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))

	fmt.Printf("\n📊 Summary: %d accounts · %d tweets · %d retweets · %d quotes\n\n",
		len(accounts), originals, retweets, quotes)

	// Most Active
	if totalTweets > 0 {
		fmt.Printf("%s\n", sectionStyle.Render("🔥 Most Active"))
		tweets, _ := tweetRepo.GetSince(since)

		tweetCounts := make(map[int64]int)
		for _, t := range tweets {
			tweetCounts[t.AccountID]++
		}

		type accountTweets struct {
			username string
			count    int
		}
		var sorted []accountTweets
		for accID, count := range tweetCounts {
			if acc, ok := accountMap[accID]; ok {
				sorted = append(sorted, accountTweets{acc.Username, count})
			}
		}

		// Simple sort
		for i := 0; i < len(sorted); i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j].count > sorted[i].count {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}

		limit := 5
		if len(sorted) < limit {
			limit = len(sorted)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("  %-20s %d tweets\n",
				userStyle.Render("@"+sorted[i].username),
				sorted[i].count)
		}
		fmt.Println()
	}

	// Most Amplified
	amplified, _ := tweetRepo.GetMostAmplified(since, 5)
	if len(amplified) > 0 {
		fmt.Printf("%s\n", sectionStyle.Render("🔁 Most Amplified"))
		for _, a := range amplified {
			fmt.Printf("  %-20s %s\n",
				userStyle.Render("@"+a.Username),
				dimStyle.Render(fmt.Sprintf("(%d times)", a.Count)))
		}
		fmt.Println()
	}

	// Notable Tweets
	topTweets, _ := tweetRepo.GetTopTweets(since, 3)
	if len(topTweets) > 0 {
		fmt.Printf("%s\n", sectionStyle.Render("💬 Notable Tweets"))
		for _, t := range topTweets {
			if acc, ok := accountMap[t.AccountID]; ok {
				content := t.Content
				if len(content) > 80 {
					content = content[:77] + "..."
				}
				fmt.Printf("  %s: %s\n", userStyle.Render("@"+acc.Username), content)
				fmt.Printf("    %s\n", dimStyle.Render(fmt.Sprintf("↳ %d likes · %d RTs", t.Likes, t.Retweets)))
			}
		}
		fmt.Println()
	}

	return nil
}
