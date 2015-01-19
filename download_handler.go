package main

import (
	"database/sql"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func DownloadHandler(cfg Config, w http.ResponseWriter, r *http.Request, db *sql.DB) {

	whitelist := []string{"image/gif", "image/png", "image/jpeg", "image/bmp", "application/pdf", "text/plain"}
	/* Security checks */
	request_id_t := strings.TrimSpace(r.FormValue("id"))
	if len(request_id_t) == 0 {
		http.Redirect(w, r, cfg.Main.Domain, 303)
		log.Printf("[LOG] No request passed\n")
		return
	}

	/* Remove leading slash if passed via arg */
	if request_id_t[0] == '/' {
		request_id_t = request_id_t[1:]
	}

	request_commands := strings.Split(request_id_t, "/")
	request_id := strings.Split(request_commands[0], ".")[0]
	if len(request_id) == 0 {
		http.Redirect(w, r, cfg.Main.Domain, 303)
		log.Printf("[LOG] Invalid format for download request\n")
		return
	}

	/*dbstmt, err := db.Prepare("SELECT name, size, diskid, uploaded FROM files WHERE fileid = '?'")
	if err != nil {
		panic(err)
	}
	defer dbstmt.Close()

	dbrow := db.QueryRow(request_id)
	*/
	dbrow := db.QueryRow("SELECT name, size, diskid, uploaded FROM files WHERE fileid = ?", request_id)
	var filename string
	var filesize int64
	var diskid string
	var uploaded int64

	err := dbrow.Scan(&filename, &filesize, &diskid, &uploaded)
	if err != nil {
		http.Redirect(w, r, cfg.Main.Domain, 303)
		log.Printf("[LOG] Requested file does not exist\n")
		return
	}

	out, err := exec.Command("file", "-biL", cfg.Directories.Upload+diskid).Output()

	// Tells IE not to try to guess the content type
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Tells the browser to expect a file
	w.Header().Set("Content-Description", "File Transfer")

	// Enables some XSS protection in IE
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	if StringInArray(strings.Split(string(out), ";")[0], whitelist) && err == nil {
		w.Header().Set("Content-Disposition", "inline; filename=\""+string(filename)+"\"")
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+string(filename)+"\"")
	}

	// watfile links currently don't expire, so tell the browser
	w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
	w.Header().Set("Cache-Control", "max-age=31536000, must-revalidate")

	w.Header().Set("Last-Modified", time.Unix(uploaded, 0).Format("Mon, 2 Jan 2006 15:04:05 MST"))
	w.Header().Set("Content-Length", strconv.FormatInt(filesize, 10))
	if cfg.Toggles.UseXaccel {
		w.Header().Set("X-Accel-Redirect", "/protected/"+diskid)
		w.Header().Set("Content-Transfer-Encoding", "binary")
	} else {
		http.ServeFile(w, r, cfg.Directories.Upload+diskid)
	}
	log.Printf("[LOG] File %s dowloaded\n", request_id)
	return
}
