// internal/account/repository.go
package account

import (
	"time"

	"github.com/jpequegn/xmon/internal/database"
)

type Account struct {
	ID          int64
	UserID      string
	Username    string
	Name        string
	Bio         string
	Followers   int
	AddedAt     time.Time
	LastFetched *time.Time
}

type Repository struct {
	db *database.DB
}

func NewRepository(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Add(userID, username, name, bio string, followers int) error {
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO accounts (user_id, username, name, bio, followers) VALUES (?, ?, ?, ?, ?)`,
		userID, username, name, bio, followers,
	)
	return err
}

func (r *Repository) Remove(username string) error {
	_, err := r.db.Exec(`DELETE FROM accounts WHERE username = ?`, username)
	return err
}

func (r *Repository) List() ([]Account, error) {
	rows, err := r.db.Query(`SELECT id, user_id, username, name, bio, followers, added_at, last_fetched FROM accounts ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.Name, &a.Bio, &a.Followers, &a.AddedAt, &a.LastFetched); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

func (r *Repository) Get(username string) (*Account, error) {
	var a Account
	err := r.db.QueryRow(
		`SELECT id, user_id, username, name, bio, followers, added_at, last_fetched FROM accounts WHERE username = ?`,
		username,
	).Scan(&a.ID, &a.UserID, &a.Username, &a.Name, &a.Bio, &a.Followers, &a.AddedAt, &a.LastFetched)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) GetByID(id int64) (*Account, error) {
	var a Account
	err := r.db.QueryRow(
		`SELECT id, user_id, username, name, bio, followers, added_at, last_fetched FROM accounts WHERE id = ?`,
		id,
	).Scan(&a.ID, &a.UserID, &a.Username, &a.Name, &a.Bio, &a.Followers, &a.AddedAt, &a.LastFetched)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) Exists(username string) bool {
	var count int
	r.db.QueryRow(`SELECT COUNT(*) FROM accounts WHERE username = ?`, username).Scan(&count)
	return count > 0
}

func (r *Repository) UpdateLastFetched(id int64) error {
	_, err := r.db.Exec(`UPDATE accounts SET last_fetched = CURRENT_TIMESTAMP WHERE id = ?`, id)
	return err
}
