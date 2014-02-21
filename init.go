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
func Init(cfg Config) *sql.DB {

	runtime.GOMAXPROCS(runtime.NumCPU() + 1)
	rand.Seed(time.Now().UTC().UnixNano())

	perm := os.ModeDir | 0755

	exists, _ := Exists(cfg.Directories.Data)
	if exists == false {
		os.Mkdir(cfg.Directories.Data, perm)
		log.Printf("[LOG] Initializing data directory")
	}
	exists, _ = Exists(cfg.Directories.Upload)
	if exists == false {
		os.Mkdir(cfg.Directories.Upload, perm)
		log.Printf("[LOG] Initializing upload directory")
	}

	db, err := sql.Open("mysql", cfg.Database.Username+":"+cfg.Database.Password+"@"+cfg.Database.Host+"/"+cfg.Database.Name)
	if err != nil {
		panic(err)
	}

	return db
}
