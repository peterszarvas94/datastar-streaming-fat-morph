package main

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

const sqlitePath = "data.sqlite"

var db *sql.DB

func initDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA synchronous=NORMAL;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec("PRAGMA busy_timeout=1000;"); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS counter_events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			client_id TEXT NOT NULL,
			action TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);
	`); err != nil {
		return nil, err
	}
	return db, nil
}
