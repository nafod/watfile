package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (

	/* Format "IP:Port number" */
	CONF_IP = ":31114"

	/* Used for redirects */
	CONF_DOMAIN = "http://watfile.com/"

	CONF_MAX_FILESIZE = 10 << 20

	/* Base watfile data directories */
	DATA_DIR    = "./data-watfile"
	UPLOAD_DIR  = DATA_DIR + "/uploads/"
	HASH_DIR    = DATA_DIR + "/hashes/"
	ACCOUNT_DIR = DATA_DIR + "/accounts/"
	DELETE_DIR  = DATA_DIR + "/delete/"
	FORCEDL_DIR = DATA_DIR + "/forcedl/"
)

func WriteFileSafe(path string, content []byte) bool {
	ioutil.WriteFile(path, content, os.ModePerm)
	return true
}

func WriteEmptyFile(path string) bool {
	ioutil.WriteFile(path, []byte{}, os.ModePerm)
	return true
}

func main() {

	/* Create initial directories, sets GOMAXPROC, and seeds the PRNG */
	Init()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		UploadHandler(w, r)
	})

	http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		FileHandler(w, r)
	})

	http.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) {
		DownloadHandler(w, r)
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		DeleteHandler(w, r)
	})

	/* API paths */
	//http.HandleFunc("/api/v1/upload", func(w http.ResponseWriter, r *http.Request) {
	//	APIUploadHandler(w, r)
	//})

	http.HandleFunc("/api/v1/dl", func(w http.ResponseWriter, r *http.Request) {
		APIDownloadHandler(w, r)
	})

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

	log.Fatal(http.ListenAndServe(CONF_IP, nil))
    log.Printf("[LOG] Now listening on %s", CONF_IP)
}
