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
