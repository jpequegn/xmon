// internal/tweet/repository.go
package tweet

import (
	"time"

	"github.com/jpequegn/xmon/internal/database"
)

type Tweet struct {
	ID               int64
	AccountID        int64
	TweetID          string
	TweetType        string
	Content          string
	ReferencedUser   string
	ReferencedTweetID string
	Likes            int
	Retweets         int
	CreatedAt        time.Time
}

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Add(accountID int64, tweetID, tweetType, content, refUser, refTweetID string, likes, retweets int, createdAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT OR IGNORE INTO tweets (account_id, tweet_id, tweet_type, content, referenced_user, referenced_tweet_id, likes, retweets, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		accountID, tweetID, tweetType, content, refUser, refTweetID, likes, retweets, createdAt,
	)
	return err
}

func (r *Repository) GetSince(since time.Time) ([]Tweet, error) {
	rows, err := r.db.Query(`
		SELECT id, account_id, tweet_id, tweet_type, content, referenced_user, referenced_tweet_id, likes, retweets, created_at
		FROM tweets
		WHERE created_at >= ?
		ORDER BY created_at DESC
	`, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []Tweet
	for rows.Next() {
		var t Tweet
		if err := rows.Scan(&t.ID, &t.AccountID, &t.TweetID, &t.TweetType, &t.Content, &t.ReferencedUser, &t.ReferencedTweetID, &t.Likes, &t.Retweets, &t.CreatedAt); err != nil {
			return nil, err
		}
		tweets = append(tweets, t)
	}
	return tweets, rows.Err()
}

func (r *Repository) GetForAccount(accountID int64, since time.Time) ([]Tweet, error) {
	rows, err := r.db.Query(`
		SELECT id, account_id, tweet_id, tweet_type, content, referenced_user, referenced_tweet_id, likes, retweets, created_at
		FROM tweets
		WHERE account_id = ? AND created_at >= ?
		ORDER BY created_at DESC
	`, accountID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []Tweet
	for rows.Next() {
		var t Tweet
		if err := rows.Scan(&t.ID, &t.AccountID, &t.TweetID, &t.TweetType, &t.Content, &t.ReferencedUser, &t.ReferencedTweetID, &t.Likes, &t.Retweets, &t.CreatedAt); err != nil {
			return nil, err
		}
		tweets = append(tweets, t)
	}
	return tweets, rows.Err()
}

func (r *Repository) CountByType(since time.Time) (originals, retweets, quotes int, err error) {
	row := r.db.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN tweet_type = 'original' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tweet_type = 'retweet' THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN tweet_type = 'quote' THEN 1 ELSE 0 END), 0)
		FROM tweets WHERE created_at >= ?
	`, since)
	err = row.Scan(&originals, &retweets, &quotes)
	return
}

func (r *Repository) GetMostAmplified(since time.Time, limit int) ([]struct {
	Username string
	Count    int
}, error) {
	rows, err := r.db.Query(`
		SELECT referenced_user, COUNT(*) as count
		FROM tweets
		WHERE created_at >= ? AND tweet_type IN ('retweet', 'quote') AND referenced_user != ''
		GROUP BY referenced_user
		ORDER BY count DESC
		LIMIT ?
	`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []struct {
		Username string
		Count    int
	}
	for rows.Next() {
		var r struct {
			Username string
			Count    int
		}
		if err := rows.Scan(&r.Username, &r.Count); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (r *Repository) GetTopTweets(since time.Time, limit int) ([]Tweet, error) {
	rows, err := r.db.Query(`
		SELECT id, account_id, tweet_id, tweet_type, content, referenced_user, referenced_tweet_id, likes, retweets, created_at
		FROM tweets
		WHERE created_at >= ? AND tweet_type = 'original'
		ORDER BY (likes + retweets) DESC
		LIMIT ?
	`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tweets []Tweet
	for rows.Next() {
		var t Tweet
		if err := rows.Scan(&t.ID, &t.AccountID, &t.TweetID, &t.TweetType, &t.Content, &t.ReferencedUser, &t.ReferencedTweetID, &t.Likes, &t.Retweets, &t.CreatedAt); err != nil {
			return nil, err
		}
		tweets = append(tweets, t)
	}
	return tweets, rows.Err()
}
