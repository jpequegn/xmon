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
