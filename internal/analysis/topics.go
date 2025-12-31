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
