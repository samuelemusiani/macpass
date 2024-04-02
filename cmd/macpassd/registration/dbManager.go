package registration

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func dbConnect(dbPath string) *sql.DB {
	slog.Info("Connecting to DB")

	// If the path is prefixed with ":" the db is created in memory. This helps
	// on tests
	if !strings.HasPrefix(dbPath, ":") {
		_, err := os.Stat(dbPath)

		if os.IsNotExist(err) {
			slog.With("dbPath", dbPath).
				Info("Creating db as not present in the path specified")

			_, err := os.Create(dbPath)
			if err != nil {
				slog.With("dbPath", dbPath).Error("Creating database file")
				return nil
			}
			slog.With("dbPath", dbPath).Info("Database file created")
		} else {
			slog.With("dbPath", dbPath).Info("Databse file is already present in path")
		}
	}

	db, err := sql.Open("sqlite3", dbPath)

	if err != nil {
		slog.With("dbPath", dbPath, "db", db, "err", err).
			Error("Opening database")

		return nil
	}

	const query = `create table if not exists Log(id integer primary key 
  autoincrement, User varchar(255), MAC varchar(17), IPs blob, 
  startTime datetime, endTime datetime, isDown bool)`

	_, err = db.Exec(query)
	if err != nil {
		slog.With("db", db, "query", query, "err", err).
			Error("Creating table on databese. Databse can't be used")

		return nil
	}

	slog.Info("Successfully connected to db")
	return db
}

// This function change the value "isDown" to true for every items in the
// db that match the ids of the items in r
func dbSetOutdated(db *sql.DB, r []Registration) {
	for i := range r {
		query := `update Log set isDown = true where id = ?`

		_, err := db.Exec(query, r[i].Id)
		if err != nil {
			slog.With("db", db, "query", query, "reg", r[i], "err", err)
		}
	}
}

func dbInsertRegistration(db *sql.DB, r Registration) {
	const query = `insert into Log(User, MAC, IPs, startTime, endTime, isDown) values(?, ?, ?, ?, ?, false)`
	_, err := db.Exec(query, r.User, r.Mac, convertIPsToBytes(r.Ips), r.Start, r.End)
	if err != nil {
		slog.With("db", db, "query", query, "registration", r.String(), "err", err).
			Error("Inserting registration into database")
	}
}

func dbGetActive(db *sql.DB) []Registration {
	selectOutdated := `select * from Log where ? < endTime and isDown = false`
	row, err := db.Query(selectOutdated, time.Now())
	if err != nil {
		slog.With("db", db, "err", err).
			Error("During query for active registrations on db")

		return make([]Registration, 0)
	}
	return dbParseRows(row)
}

// Return all registrations that are outdated based on endTime, BUT NOT FROM THE
// "isDown" value. After this, the function setOutdated(...) should be called.
func dbGetOutdated(db *sql.DB) []Registration {
	selectOutdated := `select * from Log where ? > endTime and isDown = false`
	row, err := db.Query(selectOutdated, time.Now())
	if err != nil {
		slog.With("db", db, "err", err).
			Error("During query for outdated registrations on db")

		return make([]Registration, 0)
	}
	return dbParseRows(row)
}

func dbParseRows(rows *sql.Rows) (r []Registration) {
	for rows.Next() {
		var tmp Registration
		tmpIps := make([]byte, 0)
		rows.Scan(&tmp.Id, &tmp.User, &tmp.Mac, &tmpIps, &tmp.Start, &tmp.End,
			&tmp.IsDown)

		var err error
		tmp.Ips, err = convertBytesToIPs(tmpIps)
		if err != nil {
			fmt.Println(tmpIps)
			slog.With("bytes", tmpIps, "err", err).
				Error("Could not parse ips from bytes blob extracted from database")

			tmp.Ips = nil
		}

		r = append(r, tmp)
	}
	return r
}

func convertIPsToBytes(ips []net.IP) []byte {
	blob := make([]byte, 16*len(ips))

	for i := range ips {
		tmp := []byte(ips[i])

		for j := 0; j < 16; j++ {
			blob[i*16+j] = tmp[j]
		}
	}

	return blob
}

func convertBytesToIPs(blob []byte) ([]net.IP, error) {
	l := len(blob)

	if l%16 != 0 {
		return nil, fmt.Errorf("Could not convert []byte to []net.IP. Lenght is not divisible by 16. length: %d", l)
	}

	ips := make([]net.IP, l/16)

	for i := 0; i < l; i += 16 {
		tmp := make([]byte, 16)

		for j := 0; j < 16; j++ {
			tmp[j] = blob[i+j]
		}

		ips[i/16] = net.IP(tmp)
	}

	return ips, nil
}
