// internal/x/client.go
package x

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const baseURL = "https://api.twitter.com/2"

type Client struct {
	bearerToken        string
	httpClient         *http.Client
	rateLimitRemaining int
	rateLimitReset     time.Time
}

type User struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PublicMetrics struct {
		FollowersCount int `json:"followers_count"`
		FollowingCount int `json:"following_count"`
		TweetCount     int `json:"tweet_count"`
	} `json:"public_metrics"`
}

type Tweet struct {
	ID                string    `json:"id"`
	Text              string    `json:"text"`
	CreatedAt         time.Time `json:"created_at"`
	PublicMetrics     struct {
		RetweetCount int `json:"retweet_count"`
		LikeCount    int `json:"like_count"`
	} `json:"public_metrics"`
	ReferencedTweets []struct {
		Type string `json:"type"`
		ID   string `json:"id"`
	} `json:"referenced_tweets"`
}

type UserResponse struct {
	Data User `json:"data"`
}

type TweetsResponse struct {
	Data []Tweet `json:"data"`
	Meta struct {
		NewestID    string `json:"newest_id"`
		OldestID    string `json:"oldest_id"`
		ResultCount int    `json:"result_count"`
		NextToken   string `json:"next_token"`
	} `json:"meta"`
	Includes struct {
		Users []User `json:"users"`
	} `json:"includes"`
}

func NewClient(bearerToken string) *Client {
	return &Client{
		bearerToken: bearerToken,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse rate limit headers
	if remaining := resp.Header.Get("x-rate-limit-remaining"); remaining != "" {
		if val, err := strconv.Atoi(remaining); err == nil {
			c.rateLimitRemaining = val
		}
	}
	if reset := resp.Header.Get("x-rate-limit-reset"); reset != "" {
		if val, err := strconv.ParseInt(reset, 10, 64); err == nil {
			c.rateLimitReset = time.Unix(val, 0)
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("X API error %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

func (c *Client) GetUser(username string) (*User, error) {
	url := fmt.Sprintf("%s/users/by/username/%s?user.fields=description,public_metrics", baseURL, username)
	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var resp UserResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp.Data, nil
}

func (c *Client) GetUserTweets(userID string, sinceID string) (*TweetsResponse, error) {
	url := fmt.Sprintf("%s/users/%s/tweets?max_results=100&tweet.fields=created_at,public_metrics,referenced_tweets&expansions=referenced_tweets.id.author_id&user.fields=username",
		baseURL, userID)

	if sinceID != "" {
		url += "&since_id=" + sinceID
	}

	data, err := c.doRequest(url)
	if err != nil {
		return nil, err
	}

	var resp TweetsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *Client) RateLimitRemaining() int {
	return c.rateLimitRemaining
}

func (c *Client) RateLimitReset() time.Time {
	return c.rateLimitReset
}

func (c *Client) WaitForRateLimit() {
	if c.rateLimitRemaining > 0 && c.rateLimitRemaining < 5 && time.Now().Before(c.rateLimitReset) {
		waitTime := time.Until(c.rateLimitReset) + time.Second
		fmt.Printf("Rate limit low (%d remaining), waiting %v...\n", c.rateLimitRemaining, waitTime.Round(time.Second))
		time.Sleep(waitTime)
	}
}

// GetTweetType determines if a tweet is original, retweet, or quote
func GetTweetType(tweet Tweet) string {
	for _, ref := range tweet.ReferencedTweets {
		if ref.Type == "retweeted" {
			return "retweet"
		}
		if ref.Type == "quoted" {
			return "quote"
		}
	}
	return "original"
}
