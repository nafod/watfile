package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
    "log"
	"os"
	"runtime"
	"time"
)

/*
	Initial code to set up watfile directories,
	seed PRNG, and set the number of processes
*/
func Init() *sql.DB {

	runtime.GOMAXPROCS(runtime.NumCPU() + 1)
	rand.Seed(time.Now().UTC().UnixNano())

	perm := os.ModeDir | 0755

	exists, _ := Exists(DATA_DIR)
	if exists == false {
		os.Mkdir(DATA_DIR, perm)
		log.Printf("[LOG] Initializing data directory")
	}
	exists, _ = Exists(UPLOAD_DIR)
	if exists == false {
		os.Mkdir(UPLOAD_DIR, perm)
		log.Printf("[LOG] Initializing upload directory")
	}

	db, err := sql.Open("mysql", CONF_DB_USERNAME+":"+CONF_DB_PASSWORD+"@"+CONF_DB_HOST+"/"+CONF_DB_NAME)
	if err != nil {
		panic(err)
	}

	return db
}
