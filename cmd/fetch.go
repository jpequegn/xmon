// cmd/fetch.go
package cmd

import (
	"fmt"

	"github.com/jpequegn/xmon/internal/account"
	"github.com/jpequegn/xmon/internal/config"
	"github.com/jpequegn/xmon/internal/database"
	"github.com/jpequegn/xmon/internal/tweet"
	"github.com/jpequegn/xmon/internal/x"
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch tweets from monitored accounts",
	Long:  `Downloads recent tweets from all monitored X accounts.`,
	RunE:  runFetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}

func runFetch(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.X.BearerToken == "" {
		return fmt.Errorf("X API bearer token not set. Add it to %s", config.ConfigPath())
	}

	db, err := database.New(config.DBPath())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	accountRepo := account.NewRepository(db)
	tweetRepo := tweet.NewRepository(db)

	accounts, err := accountRepo.List()
	if err != nil {
		return fmt.Errorf("failed to list accounts: %w", err)
	}

	if len(accounts) == 0 {
		fmt.Println("No accounts to fetch. Run 'xmon add <username>' first.")
		return nil
	}

	client := x.NewClient(cfg.X.BearerToken)

	fmt.Printf("Fetching tweets for %d accounts...\n\n", len(accounts))

	totalTweets := 0

	for _, acc := range accounts {
		client.WaitForRateLimit()

		tweetsResp, err := client.GetUserTweets(acc.UserID, "")
		if err != nil {
			fmt.Printf("  @%s: error - %v\n", acc.Username, err)
			continue
		}

		count := 0
		for _, tw := range tweetsResp.Data {
			tweetType := x.GetTweetType(tw)

			// Get referenced user for RTs/quotes
			refUser := ""
			refTweetID := ""
			if len(tw.ReferencedTweets) > 0 {
				refTweetID = tw.ReferencedTweets[0].ID
				// Try to find the author in includes
				for _, u := range tweetsResp.Includes.Users {
					refUser = u.Username
					break
				}
			}

			err := tweetRepo.Add(
				acc.ID,
				tw.ID,
				tweetType,
				tw.Text,
				refUser,
				refTweetID,
				tw.PublicMetrics.LikeCount,
				tw.PublicMetrics.RetweetCount,
				tw.CreatedAt,
			)
			if err == nil {
				count++
			}
		}

		accountRepo.UpdateLastFetched(acc.ID)
		fmt.Printf("  @%s: %d tweets\n", acc.Username, count)
		totalTweets += count
	}

	fmt.Printf("\nFetch complete: %d new tweets\n", totalTweets)

	if client.RateLimitRemaining() > 0 {
		fmt.Printf("Rate limit: %d requests remaining (resets %s)\n",
			client.RateLimitRemaining(),
			client.RateLimitReset().Format("15:04"))
	}

	fmt.Println("Run 'xmon digest' to see the summary.")

	return nil
}
