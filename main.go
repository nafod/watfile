package main

import (
	"code.google.com/p/gcfg"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Main struct {
		IP            string
		Domain        string
	}

	Database struct {
		Username string
		Password string
		Host     string
		Name     string
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
        /* Couldn't read the config file */
        panic(err)
    }

	/* Create initial directories, sets GOMAXPROC, and seeds the PRNG */
	db := Init(cfg)
	defer db.Close()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		UploadHandler(cfg, w, r, db)
	})
	/*
		http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
			FileHandler(w, r, db)
		})
	*/
	http.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) {
		DownloadHandler(cfg, w, r, db)
	})
	/*
		http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
			DeleteHandler(w, r, db)
		})
	*/
	/* API paths */
	//http.HandleFunc("/api/v1/upload", func(w http.ResponseWriter, r *http.Request) {
	//	APIUploadHandler(w, r)
	//})

	/*http.HandleFunc("/api/v1/dl", func(w http.ResponseWriter, r *http.Request) {
		APIDownloadHandler(w, r)
	})*/

	/*http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, mc)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		LogoutHandler(w, r, mc)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, mc)
	})*/

	http.HandleFunc("/mu-3f8488db-7fabdac2-b1583628-30caf91d", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "42")
	})

	log.Fatal(http.ListenAndServe(cfg.Main.IP, nil))
	log.Printf("[LOG] Now listening on %s", cfg.Main.IP)
}
