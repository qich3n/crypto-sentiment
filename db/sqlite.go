package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Create tables if they don't exist
	err = createTables()
	if err != nil {
		return err
	}

	return nil
}

func createTables() error {
	sentimentTable := `
    CREATE TABLE IF NOT EXISTS sentiment_data (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        symbol TEXT NOT NULL,
        score REAL NOT NULL,
        reddit_score REAL,
        twitter_score REAL,
        total_posts INTEGER,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

	_, err := DB.Exec(sentimentTable)
	if err != nil {
		return err
	}

	return nil
}
