package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitializeDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL CHECK(LENGTH(date) = 8),
		title TEXT NOT NULL CHECK(LENGTH(title) <= 255),
		description TEXT CHECK(LENGTH(description) <= 1000),
		comment TEXT CHECK(LENGTH(comment) <= 1000),
		repeat TEXT CHECK(LENGTH(repeat) <= 20)
	);`

	_, err = db.Exec(createTableQuery)
	if err != nil {
		return nil, err
	}

	DB = db
	return db, nil
}
