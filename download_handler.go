package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {

	whitelist := []string{"image/gif", "image/png", "image/jpeg", "image/bmp", "application/pdf", "text/plain"}
	/* Security checks */
	request_id_t := strings.TrimSpace(r.FormValue("id"))
	if len(request_id_t) == 0 {
		http.Redirect(w, r, CONF_DOMAIN, 303)
		return
	}

	request_commands := strings.Split(request_id_t, "/")
	request_id := strings.Split(request_commands[0], ".")[0]
	if len(request_id) == 0 {
		http.Redirect(w, r, CONF_DOMAIN, 303)
		return
	}

	exists, _ := Exists(UPLOAD_DIR + request_id + "/")
	if exists == false {
		http.Redirect(w, r, CONF_DOMAIN, 303)
		return
	}

	files_t, _ := ioutil.ReadDir(UPLOAD_DIR + request_id + "/")

	filename := ""
	for a := range files_t {
		if files_t[a].Name() != "." && files_t[a].Name() != ".." {
			filename = files_t[a].Name()
			break
		}
	}

	if len(filename) == 0 {
		http.Redirect(w, r, CONF_DOMAIN, 303)
		return
	}

	fileinfo_t, err := os.Stat(UPLOAD_DIR + request_id + "/" + filename)
	if err != nil {
		panic(err)
	}
	out, err := exec.Command("file", "-biL", UPLOAD_DIR+request_id+"/"+filename).Output()

	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	exists, _ = Exists(FORCEDL_DIR + request_id)
	base64_t, _ := base64.URLEncoding.DecodeString(filename)

	if StringInArray(strings.Split(string(out), ";")[0], whitelist) && !exists && err == nil {
		w.Header().Set("Content-Disposition", "inline; filename=\""+string(base64_t)+"\"")
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+string(base64_t)+"\"")
	}

	w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
	w.Header().Set("Cache-Control", "max-age=31536000, must-revalidate")
	w.Header().Set("Last-Modified", fileinfo_t.ModTime().Format("Mon, 2 Jan 2006 15:04:05 MST"))
	w.Header().Set("Content-Length", strconv.FormatInt(fileinfo_t.Size(), 10))
	if CONF_USE_XACCEL {
		w.Header().Set("X-Accel-Redirect", "/protected/"+request_id+"/"+filename)
		w.Header().Set("Content-Transfer-Encoding", "binary")
	} else {
		http.ServeFile(w, r, UPLOAD_DIR+request_id+"/"+filename)
	}
	log.Printf("[LOG] File %s dowloaded\n", request_id)
	return
}
