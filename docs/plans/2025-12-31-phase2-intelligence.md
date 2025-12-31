# Phase 2 Intelligence Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add LLM-powered analysis with `--smart` flag, enhanced amplification detection showing who retweeted whom, topic extraction from tweet content, and monthly API usage tracking with warnings.

**Architecture:** LLM client connects to local Ollama for generating insights from tweet data. Analysis module extracts topics/keywords from tweet content. Tweet repository enhanced to show which accounts amplified each user. API usage repository tracks monthly tweet reads against X API free tier limit (1,500/month).

**Tech Stack:** Go, Ollama API, SQLite, existing xmon packages

---

## Task 1: LLM Client Package

**Files:**
- Create: `internal/llm/client.go`
- Create: `internal/llm/client_test.go`

**Step 1: Create the LLM client**

```go
// internal/llm/client.go
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

type DigestData struct {
	TotalTweets    int
	TotalOriginal  int
	TotalRetweets  int
	TotalQuotes    int
	MostActive     []UserActivity
	MostAmplified  []AmplifiedUser
	TopTopics      []string
	NotableTweets  []NotableTweet
}

type UserActivity struct {
	Username string
	Count    int
}

type AmplifiedUser struct {
	Username   string
	AmplifiedBy []string
}

type NotableTweet struct {
	Author  string
	Content string
	Likes   int
	RTs     int
}

type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type generateResponse struct {
	Response string `json:"response"`
}

func NewClient(baseURL, model string) *Client {
	return &Client{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := generateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	var result generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return strings.TrimSpace(result.Response), nil
}

func GenerateDigestPrompt(data DigestData) string {
	var sb strings.Builder

	sb.WriteString("Analyze this X/Twitter activity digest and provide 2-3 brief insights about emerging themes and what these influential people are focusing on.\n\n")
	sb.WriteString("Activity Summary:\n")
	sb.WriteString(fmt.Sprintf("- %d total tweets (%d original, %d retweets, %d quotes)\n",
		data.TotalTweets, data.TotalOriginal, data.TotalRetweets, data.TotalQuotes))

	if len(data.TopTopics) > 0 {
		sb.WriteString(fmt.Sprintf("- Trending topics: %s\n", strings.Join(data.TopTopics, ", ")))
	}

	if len(data.MostAmplified) > 0 {
		sb.WriteString("\nMost amplified accounts (who multiple people are retweeting):\n")
		for _, a := range data.MostAmplified {
			sb.WriteString(fmt.Sprintf("- @%s amplified by: %s\n", a.Username, strings.Join(a.AmplifiedBy, ", ")))
		}
	}

	if len(data.MostActive) > 0 {
		sb.WriteString("\nMost active accounts:\n")
		for _, u := range data.MostActive {
			sb.WriteString(fmt.Sprintf("- @%s: %d tweets\n", u.Username, u.Count))
		}
	}

	if len(data.NotableTweets) > 0 {
		sb.WriteString("\nNotable tweets (highest engagement):\n")
		for _, t := range data.NotableTweets {
			content := t.Content
			if len(content) > 100 {
				content = content[:97] + "..."
			}
			sb.WriteString(fmt.Sprintf("- @%s: \"%s\" (%d likes, %d RTs)\n", t.Author, content, t.Likes, t.RTs))
		}
	}

	sb.WriteString("\nProvide 2-3 concise bullet points about emerging themes, sentiment shifts, or notable patterns. Focus on what these influential people are signaling. Keep each bullet under 120 characters.")

	return sb.String()
}
```

**Step 2: Create basic test**

```go
// internal/llm/client_test.go
package llm

import (
	"strings"
	"testing"
)

func TestGenerateDigestPrompt(t *testing.T) {
	data := DigestData{
		TotalTweets:   100,
		TotalOriginal: 50,
		TotalRetweets: 30,
		TotalQuotes:   20,
		TopTopics:     []string{"AI", "crypto", "startups"},
		MostActive: []UserActivity{
			{Username: "pmarca", Count: 25},
		},
		MostAmplified: []AmplifiedUser{
			{Username: "elonmusk", AmplifiedBy: []string{"pmarca", "naval"}},
		},
	}

	prompt := GenerateDigestPrompt(data)

	if !strings.Contains(prompt, "100 total tweets") {
		t.Error("prompt should contain tweet count")
	}
	if !strings.Contains(prompt, "AI") {
		t.Error("prompt should contain topics")
	}
	if !strings.Contains(prompt, "elonmusk") {
		t.Error("prompt should contain amplified users")
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:11434", "llama3.2")
	if client.baseURL != "http://localhost:11434" {
		t.Error("baseURL not set correctly")
	}
	if client.model != "llama3.2" {
		t.Error("model not set correctly")
	}
}
```

**Step 3: Build and test**

Run:
```bash
cd /Users/julienpequegnot/Code/xmon && go test ./internal/llm/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
cd /Users/julienpequegnot/Code/xmon && git add . && git commit -m "feat: add LLM client for Ollama integration"
```

---

## Task 2: Topic Extraction Analysis

**Files:**
- Create: `internal/analysis/topics.go`
- Create: `internal/analysis/topics_test.go`

**Step 1: Create topic extraction module**

```go
// internal/analysis/topics.go
package analysis

import (
	"regexp"
	"sort"
	"strings"
)

// Common words to filter out
var stopWords = map[string]bool{
	"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
	"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
	"with": true, "by": true, "from": true, "is": true, "are": true, "was": true,
	"were": true, "be": true, "been": true, "being": true, "have": true, "has": true,
	"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
	"could": true, "should": true, "may": true, "might": true, "must": true,
	"this": true, "that": true, "these": true, "those": true, "it": true,
	"its": true, "i": true, "you": true, "he": true, "she": true, "we": true,
	"they": true, "me": true, "him": true, "her": true, "us": true, "them": true,
	"my": true, "your": true, "his": true, "our": true, "their": true,
	"what": true, "which": true, "who": true, "whom": true, "when": true,
	"where": true, "why": true, "how": true, "all": true, "each": true,
	"every": true, "both": true, "few": true, "more": true, "most": true,
	"other": true, "some": true, "such": true, "no": true, "not": true,
	"only": true, "own": true, "same": true, "so": true, "than": true,
	"too": true, "very": true, "just": true, "can": true, "now": true,
	"new": true, "like": true, "get": true, "got": true, "going": true,
	"about": true, "into": true, "over": true, "after": true, "before": true,
	"between": true, "under": true, "again": true, "then": true, "here": true,
	"there": true, "also": true, "even": true, "still": true, "through": true,
	"rt": true, "via": true, "amp": true, "https": true, "http": true,
}

type TopicCount struct {
	Topic string
	Count int
}

// ExtractHashtags extracts hashtags from tweets
func ExtractHashtags(tweets []string) []TopicCount {
	counts := make(map[string]int)
	hashtagRe := regexp.MustCompile(`#(\w+)`)

	for _, tweet := range tweets {
		matches := hashtagRe.FindAllStringSubmatch(tweet, -1)
		for _, match := range matches {
			if len(match) > 1 {
				tag := strings.ToLower(match[1])
				counts[tag]++
			}
		}
	}

	return sortTopics(counts)
}

// ExtractKeywords extracts significant keywords from tweets
func ExtractKeywords(tweets []string, minLength int, minCount int) []TopicCount {
	counts := make(map[string]int)
	wordRe := regexp.MustCompile(`\b[a-zA-Z]{3,}\b`)

	for _, tweet := range tweets {
		// Remove URLs
		tweet = regexp.MustCompile(`https?://\S+`).ReplaceAllString(tweet, "")
		// Remove mentions
		tweet = regexp.MustCompile(`@\w+`).ReplaceAllString(tweet, "")

		words := wordRe.FindAllString(strings.ToLower(tweet), -1)
		seen := make(map[string]bool) // Count each word once per tweet

		for _, word := range words {
			if len(word) >= minLength && !stopWords[word] && !seen[word] {
				counts[word]++
				seen[word] = true
			}
		}
	}

	// Filter by minimum count
	filtered := make(map[string]int)
	for word, count := range counts {
		if count >= minCount {
			filtered[word] = count
		}
	}

	return sortTopics(filtered)
}

// ExtractTopics combines hashtags and keywords
func ExtractTopics(tweets []string, limit int) []string {
	hashtags := ExtractHashtags(tweets)
	keywords := ExtractKeywords(tweets, 4, 2)

	// Combine and dedupe
	seen := make(map[string]bool)
	var result []string

	// Prioritize hashtags
	for _, h := range hashtags {
		if !seen[h.Topic] && len(result) < limit {
			result = append(result, "#"+h.Topic)
			seen[h.Topic] = true
		}
	}

	// Add keywords
	for _, k := range keywords {
		if !seen[k.Topic] && len(result) < limit {
			result = append(result, k.Topic)
			seen[k.Topic] = true
		}
	}

	return result
}

func sortTopics(counts map[string]int) []TopicCount {
	var topics []TopicCount
	for topic, count := range counts {
		topics = append(topics, TopicCount{Topic: topic, Count: count})
	}

	sort.Slice(topics, func(i, j int) bool {
		return topics[i].Count > topics[j].Count
	})

	// Return top 10
	if len(topics) > 10 {
		topics = topics[:10]
	}

	return topics
}
```

**Step 2: Create tests**

```go
// internal/analysis/topics_test.go
package analysis

import (
	"testing"
)

func TestExtractHashtags(t *testing.T) {
	tweets := []string{
		"Working on #AI and #ML projects today",
		"Love the progress in #AI agents",
		"#crypto is heating up again",
	}

	hashtags := ExtractHashtags(tweets)

	if len(hashtags) < 2 {
		t.Errorf("expected at least 2 hashtags, got %d", len(hashtags))
	}

	// AI should be first (appears twice)
	if hashtags[0].Topic != "ai" {
		t.Errorf("expected 'ai' as top hashtag, got %s", hashtags[0].Topic)
	}
	if hashtags[0].Count != 2 {
		t.Errorf("expected count 2 for 'ai', got %d", hashtags[0].Count)
	}
}

func TestExtractKeywords(t *testing.T) {
	tweets := []string{
		"Building autonomous agents for enterprise",
		"Agents are the future of software",
		"Enterprise software is changing fast",
	}

	keywords := ExtractKeywords(tweets, 4, 2)

	// Should find "agents" and "enterprise" (appear 2+ times)
	found := make(map[string]bool)
	for _, k := range keywords {
		found[k.Topic] = true
	}

	if !found["agents"] {
		t.Error("expected 'agents' in keywords")
	}
	if !found["enterprise"] {
		t.Error("expected 'enterprise' in keywords")
	}
}

func TestExtractTopics(t *testing.T) {
	tweets := []string{
		"#AI is transforming everything",
		"Building with #AI agents",
		"The agents are getting smarter",
	}

	topics := ExtractTopics(tweets, 5)

	if len(topics) == 0 {
		t.Error("expected at least one topic")
	}

	// First should be #ai (hashtag, appears twice)
	if topics[0] != "#ai" {
		t.Errorf("expected '#ai' as first topic, got %s", topics[0])
	}
}
```

**Step 3: Build and test**

Run:
```bash
cd /Users/julienpequegnot/Code/xmon && go test ./internal/analysis/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
cd /Users/julienpequegnot/Code/xmon && git add . && git commit -m "feat: add topic/keyword extraction for tweet analysis"
```

---

## Task 3: Enhanced Amplification Detection

**Files:**
- Modify: `internal/tweet/repository.go`

**Step 1: Read current repository**

Read `/Users/julienpequegnot/Code/xmon/internal/tweet/repository.go` to understand current structure.

**Step 2: Add enhanced amplification method**

Add this new method to `internal/tweet/repository.go` that returns who amplified each user:

```go
// AmplifiedUser represents a user who was RTd/quoted with who amplified them
type AmplifiedUser struct {
	Username    string
	AmplifiedBy []string
	Count       int
}

// GetAmplifiedWithSources returns users who were RTd/quoted along with who amplified them
func (r *Repository) GetAmplifiedWithSources(since time.Time, minAmplifiers int) ([]AmplifiedUser, error) {
	rows, err := r.db.Query(`
		SELECT t.referenced_user, a.username
		FROM tweets t
		JOIN accounts a ON t.account_id = a.id
		WHERE t.created_at >= ?
			AND t.tweet_type IN ('retweet', 'quote')
			AND t.referenced_user != ''
		ORDER BY t.referenced_user
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Map: referenced_user -> set of amplifiers
	amplifierMap := make(map[string]map[string]bool)

	for rows.Next() {
		var refUser, amplifier string
		if err := rows.Scan(&refUser, &amplifier); err != nil {
			return nil, err
		}
		if amplifierMap[refUser] == nil {
			amplifierMap[refUser] = make(map[string]bool)
		}
		amplifierMap[refUser][amplifier] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert to slice, filter by minimum amplifiers
	var results []AmplifiedUser
	for user, amplifiers := range amplifierMap {
		if len(amplifiers) >= minAmplifiers {
			var names []string
			for name := range amplifiers {
				names = append(names, name)
			}
			results = append(results, AmplifiedUser{
				Username:    user,
				AmplifiedBy: names,
				Count:       len(names),
			})
		}
	}

	// Sort by count descending
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Count > results[i].Count {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Limit to top 10
	if len(results) > 10 {
		results = results[:10]
	}

	return results, nil
}
```

**Step 3: Build and verify**

```bash
cd /Users/julienpequegnot/Code/xmon && go build -o xmon .
```

**Step 4: Commit**

```bash
cd /Users/julienpequegnot/Code/xmon && git add . && git commit -m "feat: add enhanced amplification detection with source tracking"
```

---

## Task 4: API Usage Repository

**Files:**
- Create: `internal/usage/repository.go`
- Create: `internal/usage/repository_test.go`

**Step 1: Create usage repository**

```go
// internal/usage/repository.go
package usage

import (
	"fmt"
	"time"

	"github.com/jpequegn/xmon/internal/database"
)

const MonthlyLimit = 1500 // X API free tier limit

type Repository struct {
	db *database.DB
}

type MonthlyUsage struct {
	Month      string
	TweetsRead int
	UpdatedAt  time.Time
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

// GetCurrentMonth returns usage for the current month
func (r *Repository) GetCurrentMonth() (*MonthlyUsage, error) {
	month := time.Now().Format("2006-01")
	return r.GetMonth(month)
}

// GetMonth returns usage for a specific month
func (r *Repository) GetMonth(month string) (*MonthlyUsage, error) {
	row := r.db.QueryRow(`
		SELECT month, tweets_read, updated_at
		FROM api_usage
		WHERE month = ?
	`, month)

	var usage MonthlyUsage
	err := row.Scan(&usage.Month, &usage.TweetsRead, &usage.UpdatedAt)
	if err != nil {
		// Return zero usage if not found
		return &MonthlyUsage{Month: month, TweetsRead: 0}, nil
	}

	return &usage, nil
}

// AddTweetsRead increments the tweet count for current month
func (r *Repository) AddTweetsRead(count int) error {
	month := time.Now().Format("2006-01")

	_, err := r.db.Exec(`
		INSERT INTO api_usage (month, tweets_read, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(month) DO UPDATE SET
			tweets_read = tweets_read + ?,
			updated_at = CURRENT_TIMESTAMP
	`, month, count, count)

	return err
}

// GetRemainingQuota returns how many tweets can still be read this month
func (r *Repository) GetRemainingQuota() (int, error) {
	usage, err := r.GetCurrentMonth()
	if err != nil {
		return MonthlyLimit, err
	}

	remaining := MonthlyLimit - usage.TweetsRead
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// CheckQuota returns a warning message if quota is low, empty string otherwise
func (r *Repository) CheckQuota() string {
	remaining, _ := r.GetRemainingQuota()
	usage, _ := r.GetCurrentMonth()

	percentUsed := float64(usage.TweetsRead) / float64(MonthlyLimit) * 100

	if remaining == 0 {
		return fmt.Sprintf("⚠️  Monthly API limit reached! %d/%d tweets read (%.0f%%)",
			usage.TweetsRead, MonthlyLimit, percentUsed)
	}

	if percentUsed >= 90 {
		return fmt.Sprintf("⚠️  API quota critical: %d/%d tweets read (%.0f%%), %d remaining",
			usage.TweetsRead, MonthlyLimit, percentUsed, remaining)
	}

	if percentUsed >= 75 {
		return fmt.Sprintf("⚠️  API quota warning: %d/%d tweets read (%.0f%%), %d remaining",
			usage.TweetsRead, MonthlyLimit, percentUsed, remaining)
	}

	return ""
}
```

**Step 2: Create tests**

```go
// internal/usage/repository_test.go
package usage

import (
	"os"
	"testing"

	"github.com/jpequegn/xmon/internal/database"
)

func TestUsageRepository(t *testing.T) {
	// Create temp database
	tmpFile, err := os.CreateTemp("", "xmon-usage-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, err := database.New(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)

	// Test initial state
	usage, err := repo.GetCurrentMonth()
	if err != nil {
		t.Fatal(err)
	}
	if usage.TweetsRead != 0 {
		t.Errorf("expected 0 initial reads, got %d", usage.TweetsRead)
	}

	// Test adding reads
	err = repo.AddTweetsRead(100)
	if err != nil {
		t.Fatal(err)
	}

	usage, _ = repo.GetCurrentMonth()
	if usage.TweetsRead != 100 {
		t.Errorf("expected 100 reads, got %d", usage.TweetsRead)
	}

	// Test adding more
	repo.AddTweetsRead(50)
	usage, _ = repo.GetCurrentMonth()
	if usage.TweetsRead != 150 {
		t.Errorf("expected 150 reads, got %d", usage.TweetsRead)
	}

	// Test remaining quota
	remaining, _ := repo.GetRemainingQuota()
	if remaining != MonthlyLimit-150 {
		t.Errorf("expected %d remaining, got %d", MonthlyLimit-150, remaining)
	}
}

func TestCheckQuota(t *testing.T) {
	tmpFile, _ := os.CreateTemp("", "xmon-quota-test-*.db")
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	db, _ := database.New(tmpFile.Name())
	defer db.Close()

	repo := NewRepository(db)

	// Low usage - no warning
	repo.AddTweetsRead(500)
	warning := repo.CheckQuota()
	if warning != "" {
		t.Errorf("expected no warning at 500, got: %s", warning)
	}

	// High usage - warning
	repo.AddTweetsRead(700) // Total: 1200 (80%)
	warning = repo.CheckQuota()
	if warning == "" {
		t.Error("expected warning at 80%")
	}
}
```

**Step 3: Build and test**

```bash
cd /Users/julienpequegnot/Code/xmon && go test ./internal/usage/... -v
```

Expected: PASS

**Step 4: Commit**

```bash
cd /Users/julienpequegnot/Code/xmon && git add . && git commit -m "feat: add API usage tracking repository"
```

---

## Task 5: Integrate Usage Tracking into Fetch

**Files:**
- Modify: `cmd/fetch.go`

**Step 1: Read current fetch.go**

Read `/Users/julienpequegnot/Code/xmon/cmd/fetch.go` to understand current structure.

**Step 2: Add usage tracking**

Update `cmd/fetch.go` to:

1. Add import for usage package:
```go
"github.com/jpequegn/xmon/internal/usage"
```

2. Create usage repo after database opens:
```go
usageRepo := usage.NewRepository(db)
```

3. Check quota before fetching (after accounts check):
```go
// Check API quota
if warning := usageRepo.CheckQuota(); warning != "" {
	fmt.Println(warning)
}
```

4. Track usage after each account fetch (after the for loop processes tweets):
```go
// Track API usage
if len(tweetsResp.Data) > 0 {
	usageRepo.AddTweetsRead(len(tweetsResp.Data))
}
```

5. Show usage summary at end (before "Run 'xmon digest'"):
```go
// Show quota status
if warning := usageRepo.CheckQuota(); warning != "" {
	fmt.Println(warning)
} else {
	remaining, _ := usageRepo.GetRemainingQuota()
	fmt.Printf("API quota: %d/%d tweets remaining this month\n", remaining, usage.MonthlyLimit)
}
```

**Step 3: Build and verify**

```bash
cd /Users/julienpequegnot/Code/xmon && go build -o xmon . && ./xmon fetch --help
```

**Step 4: Commit**

```bash
cd /Users/julienpequegnot/Code/xmon && git add . && git commit -m "feat: integrate API usage tracking into fetch command"
```

---

## Task 6: Update Digest with --smart Flag

**Files:**
- Modify: `cmd/digest.go`

**Step 1: Read current digest.go**

Read `/Users/julienpequegnot/Code/xmon/cmd/digest.go` to understand current structure.

**Step 2: Add imports and flag**

Add to imports:
```go
"context"
"strings"

"github.com/jpequegn/xmon/internal/analysis"
"github.com/jpequegn/xmon/internal/llm"
```

Add flag variable after `digestDays`:
```go
var (
	digestDays  int
	digestSmart bool
)
```

In `init()`, add the flag:
```go
digestCmd.Flags().BoolVar(&digestSmart, "smart", false, "Use LLM for intelligent analysis")
```

**Step 3: Add topics section to digest**

After the "Most Amplified" section in `runDigest`, add:

```go
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
```

**Step 4: Add smart analysis at end**

At the end of `runDigest` (before `return nil`), add:

```go
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
```

**Step 5: Build and test**

```bash
cd /Users/julienpequegnot/Code/xmon && go build -o xmon . && ./xmon digest --help
```

Expected: Help shows `--smart` flag

**Step 6: Commit**

```bash
cd /Users/julienpequegnot/Code/xmon && git add . && git commit -m "feat: add --smart flag with LLM analysis to digest command"
```

---

## Task 7: Update README and Final Tests

**Files:**
- Modify: `README.md`

**Step 1: Run all tests**

```bash
cd /Users/julienpequegnot/Code/xmon && go test ./... -v
```

Expected: All tests pass

**Step 2: Update README.md**

Update the Development Status section:

```markdown
## Development Status

### Phase 1 (MVP) - Complete
- [x] Project setup (Go, Cobra, SQLite)
- [x] init command (config, database, bearer token setup)
- [x] add / remove commands
- [x] accounts command
- [x] X API client (user lookup, timeline fetch)
- [x] fetch command (pull tweets, respect rate limits)
- [x] digest command (basic aggregation)
- [x] show command

### Phase 2 (Intelligence) - Complete
- [x] LLM integration (Ollama) for --smart summaries
- [x] "Most Amplified" detection (who multiple accounts RT)
- [x] Topic/keyword extraction
- [x] Notable tweets ranking
- [x] Monthly API usage tracking with warnings

### Phase 3 (Export & Polish) - Planned
- [ ] export command (markdown generation)
- [ ] Daemon mode (scheduled fetching)
- [ ] Scraping fallback (if API limits prove too restrictive)
```

Also add to the Commands section description for --smart:

```markdown
| `xmon digest` | Show activity summary (--smart for AI insights) |
```

And add example in Quick Start:

```markdown
# Get AI-powered insights
xmon digest --smart
```

**Step 3: Commit and push**

```bash
cd /Users/julienpequegnot/Code/xmon && git add . && git commit -m "docs: mark Phase 2 Intelligence as complete"
cd /Users/julienpequegnot/Code/xmon && git push
```

---

## Summary

**Phase 2 delivers:**
- LLM client package for Ollama integration
- Topic/keyword extraction from tweet content
- Enhanced amplification detection showing who RTd whom
- API usage tracking with monthly quota warnings
- `--smart` flag on digest for AI-generated insights

**New capabilities:**
- `xmon digest --smart` - Get LLM-generated insights about themes and patterns
- `xmon fetch` - Now shows API quota status and warnings
- Trending topics section in digest showing hashtags and keywords
- "Most Amplified" section shows which accounts amplified each user
