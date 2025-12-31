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
