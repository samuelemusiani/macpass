package db

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var database *sql.DB

func Connect(sqlFile string) {
	slog.With("path", sqlFile).Debug("Connecting to DB")

	const createTable = `
  CREATE TABLE IF NOT EXISTS users(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user VARCHAR(255) UNIQUE NOT NULL,
    mac VARCHAR(17) NOT NULL
  );
  CREATE TABLE IF NOT EXISTS ssh_keys(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ssh_key TEXT UNIQUE NOT NULL,
    user_id INTEGER,
    FOREIGN KEY(user_id) REFERENCES users(id)
  );
	`
	if !strings.HasPrefix(sqlFile, ":") {
		if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
			os.Create(sqlFile)
		}
	}
	db, err := sql.Open("sqlite3", sqlFile+"?_foreign_keys=on")

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
	slog.Debug("Connected to DB")
	database = db
}

func InsertUser(user string, mac string) {
	const insertLog = `INSERT OR REPLACE INTO users(user, mac) VALUES(?, ?)`
	_, err := database.Exec(insertLog, user, mac)
	if err != nil {
		slog.With("user", user, "mac", mac, "err", err).Error("Inserting user to db")
	}
}

func AddKeyToUser(user string, key string) error {
	_, err := database.Exec(`INSERT INTO ssh_keys (ssh_key, user_id) VALUES(
    ?, (SELECT id FROM users WHERE user = ?))`, key, user)

	if err != nil {
		return err
	}

	return nil
}

func GetKeysFromUser(user string) ([]string, error) {
	rows, err := database.Query(`SELECT ssh_key FROM ssh_keys WHERE user_id = (
    SELECT id FROM users WHERE user = ?)`, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var tmp string
		err := rows.Scan(&tmp)
		if err != nil {
			return nil, err
		}

		result = append(result, tmp)
	}

	return result, nil
}

func GetMac(user string) (mac string, isPresent bool) {
	selectOutdated := `SELECT mac FROM users WHERE user=?`
	row := database.QueryRow(selectOutdated, user)

	err := row.Scan(&mac)

	if err == sql.ErrNoRows {
		log.Println("User not found in db")
		return "", false
	} else if err != nil {
		slog.With("user", user, "err", err).Error("Quering user failed")
		return "", false
	} else {
		log.Println("User FOUND in db")
		return mac, true
	}
}
