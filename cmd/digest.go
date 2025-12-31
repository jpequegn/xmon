// cmd/digest.go
package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/jpequegn/xmon/internal/account"
	"github.com/jpequegn/xmon/internal/analysis"
	"github.com/jpequegn/xmon/internal/config"
	"github.com/jpequegn/xmon/internal/database"
	"github.com/jpequegn/xmon/internal/llm"
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
	digestDays  int
	digestSmart bool
)

func init() {
	rootCmd.AddCommand(digestCmd)
	digestCmd.Flags().IntVar(&digestDays, "days", 7, "Number of days to include in digest")
	digestCmd.Flags().BoolVar(&digestSmart, "smart", false, "Use LLM for intelligent analysis")
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

	// Trending Topics
	tweets, _ := tweetRepo.GetSince(since)
	var tweetContents []string
	for _, t := range tweets {
		if t.Content != "" {
			tweetContents = append(tweetContents, t.Content)
		}
	}

	topics := analysis.ExtractTopics(tweetContents, 8)
	if len(topics) > 0 {
		fmt.Printf("%s\n", sectionStyle.Render("📢 Trending Topics"))
		fmt.Printf("  %s\n\n", dimStyle.Render(strings.Join(topics, " · ")))
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

	// Smart analysis with LLM
	if digestSmart {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("  %s\n\n", dimStyle.Render("Note: Could not load config for LLM analysis"))
		} else {
			fmt.Printf("%s\n", sectionStyle.Render("💡 Key Themes (AI-generated)"))

			// Get enhanced amplification data
			amplifiedUsers, _ := tweetRepo.GetAmplifiedWithSources(since, 2)
			var llmAmplified []llm.AmplifiedUser
			for _, a := range amplifiedUsers {
				llmAmplified = append(llmAmplified, llm.AmplifiedUser{
					Username:    a.Username,
					AmplifiedBy: a.AmplifiedBy,
				})
			}

			// Get most active
			tweetCounts := make(map[int64]int)
			for _, t := range tweets {
				tweetCounts[t.AccountID]++
			}
			var llmActive []llm.UserActivity
			for accID, count := range tweetCounts {
				if acc, ok := accountMap[accID]; ok {
					llmActive = append(llmActive, llm.UserActivity{
						Username: acc.Username,
						Count:    count,
					})
				}
			}
			// Sort by count
			for i := 0; i < len(llmActive); i++ {
				for j := i + 1; j < len(llmActive); j++ {
					if llmActive[j].Count > llmActive[i].Count {
						llmActive[i], llmActive[j] = llmActive[j], llmActive[i]
					}
				}
			}
			if len(llmActive) > 5 {
				llmActive = llmActive[:5]
			}

			// Get notable tweets
			topTweets, _ := tweetRepo.GetTopTweets(since, 3)
			var llmNotable []llm.NotableTweet
			for _, t := range topTweets {
				if acc, ok := accountMap[t.AccountID]; ok {
					llmNotable = append(llmNotable, llm.NotableTweet{
						Author:  acc.Username,
						Content: t.Content,
						Likes:   t.Likes,
						RTs:     t.Retweets,
					})
				}
			}

			digestData := llm.DigestData{
				TotalTweets:   totalTweets,
				TotalOriginal: originals,
				TotalRetweets: retweets,
				TotalQuotes:   quotes,
				TopTopics:     topics,
				MostAmplified: llmAmplified,
				MostActive:    llmActive,
				NotableTweets: llmNotable,
			}

			prompt := llm.GenerateDigestPrompt(digestData)
			client := llm.NewClient("http://localhost:11434", cfg.APIs.LLMModel)

			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			response, err := client.Generate(ctx, prompt)
			cancel()

			if err != nil {
				fmt.Printf("  %s\n\n", dimStyle.Render(fmt.Sprintf("LLM analysis unavailable: %v", err)))
			} else {
				for _, line := range strings.Split(response, "\n") {
					line = strings.TrimSpace(line)
					if line != "" {
						fmt.Printf("  %s\n", line)
					}
				}
				fmt.Println()
			}
		}
	}

	return nil
}
