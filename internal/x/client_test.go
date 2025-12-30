// internal/x/client_test.go
package x

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-token")
	if client == nil {
		t.Error("expected non-nil client")
	}
}

func TestGetTweetType(t *testing.T) {
	tests := []struct {
		name     string
		tweet    Tweet
		expected string
	}{
		{
			name:     "original tweet",
			tweet:    Tweet{},
			expected: "original",
		},
		{
			name: "retweet",
			tweet: Tweet{
				ReferencedTweets: []struct {
					Type string `json:"type"`
					ID   string `json:"id"`
				}{{Type: "retweeted", ID: "123"}},
			},
			expected: "retweet",
		},
		{
			name: "quote tweet",
			tweet: Tweet{
				ReferencedTweets: []struct {
					Type string `json:"type"`
					ID   string `json:"id"`
				}{{Type: "quoted", ID: "456"}},
			},
			expected: "quote",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetTweetType(tt.tweet)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
