// internal/account/repository_test.go
package account

import (
	"os"
	"testing"

	"github.com/jpequegn/xmon/internal/database"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	tmpfile, err := os.CreateTemp("", "xmon-test-*.db")
	if err != nil {
		t.Fatal(err)
	}

	db, err := database.New(tmpfile.Name())
	if err != nil {
		os.Remove(tmpfile.Name())
		t.Fatal(err)
	}

	return db, func() {
		db.Close()
		os.Remove(tmpfile.Name())
	}
}

func TestAddAccount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	err := repo.Add("123456", "testuser", "Test User", "A test bio", 1000)
	if err != nil {
		t.Fatalf("failed to add account: %v", err)
	}

	if !repo.Exists("testuser") {
		t.Error("account should exist after adding")
	}
}

func TestListAccounts(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	repo.Add("123", "alice", "Alice", "", 100)
	repo.Add("456", "bob", "Bob", "", 200)

	accounts, err := repo.List()
	if err != nil {
		t.Fatalf("failed to list: %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}
}

func TestRemoveAccount(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	repo.Add("123", "testuser", "Test", "", 100)
	repo.Remove("testuser")

	if repo.Exists("testuser") {
		t.Error("account should not exist after removal")
	}
}
