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
	slog.With("path", sqlFile).Info("Connecting to DB")

	const createTable = `CREATE TABLE IF NOT EXISTS 
	Inserted(User VARCHAR(255) PRIMARY KEY, mac VARCHAR(17))
	`
	if !strings.HasPrefix(sqlFile, ":") {
		if _, err := os.Stat(sqlFile); os.IsNotExist(err) {
			os.Create(sqlFile)
		}
	}
	db, err := sql.Open("sqlite3", sqlFile)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createTable)
	if err != nil {
		log.Fatal(err)
	}
	slog.Info("Connected to DB")
	database = db
}

func InsertUser(user string, mac string) {
	const insertLog = `INSERT OR REPLACE INTO Inserted(User, mac) VALUES(?, ?)`
	_, err := database.Exec(insertLog, user, mac)
	if err != nil {
		slog.With("user", user, "mac", mac, "err", err).Error("Inserting user to db")
	}
}

func GetMac(user string) (mac string, isPresent bool) {
	selectOutdated := `SELECT mac FROM Inserted WHERE user=?`
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
