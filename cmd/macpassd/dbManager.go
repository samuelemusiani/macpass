package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type registrationInstance struct {
	id     int
	user   string
	mac    string
	start  time.Time
	end    time.Time
	isDown bool
}

func regInstanceToMacRegistration(reg registrationInstance) macRegistration {
	return macRegistration{mac: reg.mac, reg: registration{reg.user, reg.start, reg.end.Sub(reg.start)}}
}

func _readRegInstances(row *sql.Rows) []registrationInstance {
	var macs []registrationInstance
	var current registrationInstance
	for row.Next() {
		row.Scan(&current.id, &current.user, &current.mac, &current.start, &current.end, &current.isDown)
		macs = append(macs, current)
	}
	return macs
}

func getActive(con *sql.DB) []registrationInstance {
	selectOutdated := `select * from Log where ? < endTime and isDown = false`
	row, err := con.Query(selectOutdated, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	return _readRegInstances(row)
}

func getOutdated(con *sql.DB) []registrationInstance {
	selectOutdated := `select * from Log where ? > endTime and isDown = false`
	row, err := con.Query(selectOutdated, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	return _readRegInstances(row)
}

func setOutdated(con *sql.DB, macs []registrationInstance) {
	for i := range macs {
		id := macs[i].id
		updateOutdated := `update Log set isDown = true where id = ?`
		con.Exec(updateOutdated, id)
	}
}

func insertMacRegistration(db *sql.DB, macReg macRegistration) {
	var startCopy = macReg.reg.start
	var endTime = startCopy.Add(macReg.reg.duration)
	const insertLog = `insert into Log(User, MAC, startTime, endTime, isDown) values(?, ?, ?,? , false)`
	_, err := db.Exec(insertLog, macReg.reg.user, macReg.mac, macReg.reg.start, endTime)
	if err != nil {
		log.Fatal(err)
	}

}

func connectDB(sqlfile string) *sql.DB {
	log.Println("Connecting to DB...")
	const createLogTable = `create table if not exists 
	Log(id integer primary key autoincrement,User varchar(255), MAC varchar(17),startTime datetime, endTime datetime, isDown bool)
	`
	if _, err := os.Stat(sqlfile); os.IsNotExist(err) {
		os.Create(sqlfile)
	}
	db, err := sql.Open("sqlite3", sqlfile)

	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createLogTable)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected to DB...")
	return db
}
