// internal/database/database.go
package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func New(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	wrapper := &DB{db}
	if err := wrapper.initSchema(); err != nil {
		return nil, err
	}

	return wrapper, nil
}

func (db *DB) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER PRIMARY KEY,
		user_id TEXT UNIQUE NOT NULL,
		username TEXT NOT NULL,
		name TEXT,
		bio TEXT,
		followers INTEGER,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_fetched DATETIME
	);

	CREATE TABLE IF NOT EXISTS tweets (
		id INTEGER PRIMARY KEY,
		account_id INTEGER NOT NULL,
		tweet_id TEXT UNIQUE NOT NULL,
		tweet_type TEXT NOT NULL,
		content TEXT,
		referenced_user TEXT,
		referenced_tweet_id TEXT,
		likes INTEGER DEFAULT 0,
		retweets INTEGER DEFAULT 0,
		created_at DATETIME,
		fetched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (account_id) REFERENCES accounts(id)
	);

	CREATE TABLE IF NOT EXISTS api_usage (
		id INTEGER PRIMARY KEY,
		month TEXT UNIQUE NOT NULL,
		tweets_read INTEGER DEFAULT 0,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_tweets_account ON tweets(account_id);
	CREATE INDEX IF NOT EXISTS idx_tweets_created ON tweets(created_at);
	CREATE INDEX IF NOT EXISTS idx_tweets_type ON tweets(tweet_type);
	`

	_, err := db.Exec(schema)
	return err
}
