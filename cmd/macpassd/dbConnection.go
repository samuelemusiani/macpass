package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type MacRegistration struct {
	mac  string
	id   int
	user string
}

func getActive(con *sql.DB) []MacRegistration {
	selectOutdated := `select id,User,MAC from Log where ? > maxTime and isDown = true`
	row, err := con.Query(selectOutdated, time.Now())
	if err != nil {
		log.Fatal(err)
	}

	var macs []MacRegistration
	var user string
	var mac string
	var id int
	for row.Next() {
		row.Scan(&id, &user, &mac)
		macs = append(macs, MacRegistration{mac, 0, user})
	}
	return macs
}
func setOutdated(con *sql.DB) []MacRegistration {
	selectOutdated := `select id,User,MAC from Log where ? < maxTime and isDown = false`
	row, err := con.Query(selectOutdated, time.Now())
	if err != nil {
		log.Fatal(err)
	}
	var macs []MacRegistration
	var user string
	var mac string
	var id int
	for row.Next() {
		row.Scan(&id, &user, &mac)
		macs = append(macs, MacRegistration{mac, 0, user})
	}
	for i := range macs {
		id := macs[i].id
		updateOutdated := `update Log set isDown = true where id = ?`
		con.Exec(updateOutdated, id)
	}
	return macs
}

func insertDB(user string, mac string, endTIme time.Time, duration time.Duration, db *sql.DB) {
	var now = time.Now()
	var maxTime = now.Add(duration)
	const insertLog = `insert into Log(User, MAC, startTime, maxTime, isDown) values(?, ?, ?,? , false)`
	db.Exec(insertLog, user, mac, now, maxTime)
}

func initDBConnection(sqlfile string) (*sql.DB, error) {
	fmt.Print("Initializing DB connection...")
	const createLogTable = `create table if not exists 
	Log(id integer primary key auto increment,User varchar(255), MAC varchar(17),startTime datetime, maxTime datetime, isDown bool)
	`
	if _, err := os.Stat(sqlfile); os.IsNotExist(err) {
		os.Create(sqlfile)
	}
	db, err := sql.Open("sqlite3", sqlfile)
	if err != nil {
		log.Fatal(err)
	}

	if err != nil {
		log.Printf("%q \n", err)
		return nil, err
	}

	_, err = db.Exec(createLogTable)
	if err != nil {
		log.Printf("%q: %s\n", err, createLogTable)
		return nil, err
	}
	return db, nil
}
