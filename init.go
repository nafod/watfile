package main

import (
	"log"
	"math/rand"
	"os"
	"runtime"
	"time"
)

/*
	Initial code to set up watfile directories,
	seed PRNG, and set the number of processes
*/
func Init() {

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
	exists, _ = Exists(HASH_DIR)
	if exists == false {
		os.Mkdir(HASH_DIR, perm)
		log.Printf("[LOG] Initializing file hash directory")
	}
	exists, _ = Exists(ACCOUNT_DIR)
	if exists == false {
		os.Mkdir(ACCOUNT_DIR, perm)
		log.Printf("[LOG] Initializing user account directory")
	}
	exists, _ = Exists(DELETE_DIR)
	if exists == false {
		os.Mkdir(DELETE_DIR, perm)
		log.Printf("[LOG] Initializing file delete metadata directory")
	}
	exists, _ = Exists(FORCEDL_DIR)
	if exists == false {
		os.Mkdir(FORCEDL_DIR, perm)
		log.Printf("[LOG] Initializing force download metadata directory")
	}
}
