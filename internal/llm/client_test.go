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
