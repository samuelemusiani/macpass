package registration

import (
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func readRegInstances(row *sql.Rows) []Registration {
	var macs []Registration
	var current Registration
	for row.Next() {
		row.Scan(&current.Id, &current.User, &current.Mac, &current.Start, &current.End, &current.IsDown)
		macs = append(macs, current)
	}
	return macs
}

func getActive(con *sql.DB) []Registration {
	selectOutdated := `select * from Log where ? < endTime and isDown = false`
	row, err := con.Query(selectOutdated, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	return readRegInstances(row)
}

func getOutdated(con *sql.DB) []Registration {
	selectOutdated := `select * from Log where ? > endTime and isDown = false`
	row, err := con.Query(selectOutdated, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	return readRegInstances(row)
}

func setOutdated(con *sql.DB, macs []Registration) {
	for i := range macs {
		id := macs[i].Id
		updateOutdated := `update Log set isDown = true where id = ?`
		con.Exec(updateOutdated, id)
		macs[i].IsDown = true
	}
}

func insertRegistration(db *sql.DB, reg Registration) {
	const insertLog = `insert into Log(User, MAC, startTime, endTime, isDown) values(?, ?, ?,? , false)`
	_, err := db.Exec(insertLog, reg.User, reg.Mac, reg.Start, reg.End)
	if err != nil {
		log.Fatal(err)
	}

}

func connectDB(sqlFile string) *sql.DB {
	log.Println("Connecting to DB")
	const createLogTable = `create table if not exists 
	Log(id integer primary key autoincrement,User varchar(255), MAC varchar(17),startTime datetime, endTime datetime, isDown bool)
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

	_, err = db.Exec(createLogTable)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to DB")
	return db
}
