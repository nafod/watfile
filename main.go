package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (

	/* Enable developer mode */
	//CONF_IP     = ":62000"
	//CONF_DOMAIN = "http://dev.watfile.com"
	//DATA_DIR    = "./dev-data-watfile"
	UPLOAD_DIR    = DATA_DIR + "/uploads/"

    CONF_DB_USERNAME = "wfproduction"
    CONF_DB_PASSWORD = "F18x2id72Xew8s9719O5Ar87v88Hcd"
    CONF_DB_HOST = "localhost"
    CONF_DB_NAME = "watfile"

	/* Format "IP:Port number" */
	CONF_IP = ":31114"

	/* Used for redirects */
	CONF_DOMAIN = "http://watfile.com/"

	/* Base watfile data directory */
	DATA_DIR = "./data-watfile"

	CONF_MAX_FILESIZE = 10 << 20

	/* Toggles */

	/* Enable nginx X-Accel */
	CONF_USE_XACCEL = true

	/* Enable Ratelimiting */
	CONF_USE_RATELIMITING = true

	/* Max files to upload in one period */
	CONF_RATELIMIT_FILES = 30

	/* Length of period in seconds */
	CONF_RATELIMIT_TIME = 300

	HASH_DIR      = DATA_DIR + "/hashes/"
	ACCOUNT_DIR   = DATA_DIR + "/accounts/"
	DELETE_DIR    = DATA_DIR + "/delete/"
	FORCEDL_DIR   = DATA_DIR + "/forcedl/"
	RATELIMIT_DIR = DATA_DIR + "/ratelimit/"
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
	db := Init()
    defer db.Close()

	http.Handle("/", http.FileServer(http.Dir("./static")))

	/*
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, "./static/index.html")
		})
		// Not needed (incorrect actually) because of the previous handler serving all / as static
	*/

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		UploadHandler(w, r, db)
	})
/*
	http.HandleFunc("/file", func(w http.ResponseWriter, r *http.Request) {
		FileHandler(w, r, db)
	})
*/
	http.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) {
		DownloadHandler(w, r, db)
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
