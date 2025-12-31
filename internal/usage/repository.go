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
