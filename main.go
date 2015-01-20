package main

import (
	"code.google.com/p/gcfg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Main struct {
		IP     string
		Domain string
	}

	Database struct {
		DSN string
	}

	Toggles struct {
		UseRatelimit bool
		UseXaccel    bool
	}

	Limits struct {
		MaxFilesize    uint64
		RatelimitFiles uint64
		RatelimitTime  uint64
	}

	Directories struct {
		Data      string
		Upload    string
		Hash      string
		Account   string
		Delete    string
		ForceDL   string
		Ratelimit string
	}
}

func WriteFileSafe(path string, content []byte) bool {
	ioutil.WriteFile(path, content, os.ModePerm)
	return true
}

func WriteEmptyFile(path string) bool {
	ioutil.WriteFile(path, []byte{}, os.ModePerm)
	return true
}

func main() {

	/* Load the configuration */
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, "watfile.conf")
	if err != nil {
		log.Panicf("[ERR] Could not read watfile.conf")
	}

	/* Create initial directories, sets GOMAXPROC, and seeds the PRNG */
	db := Init(cfg)
	defer db.Close()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		UploadHandler(cfg, w, r, db)
	})

	http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		FileHandler(cfg, w, r, db)
	})

	http.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) {
		DownloadHandler(cfg, w, r, db)
	})
	/*
		http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
			DeleteHandler(w, r, db)
		})
	*/

	log.Fatal(http.ListenAndServe(cfg.Main.IP, nil))
	log.Printf("[LOG] Now listening on %s", cfg.Main.IP)
}
