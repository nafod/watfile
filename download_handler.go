package main

import (
	"net/http"
	"strings"
	"io/ioutil"
	"os/exec"
	"os"
	"encoding/base64"
	"fmt"
	"crypto/md5"
	"crypto/sha1"
	"log"
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

	fileinfo_t, _ := os.Stat(UPLOAD_DIR + request_id + "/" + filename)
	out, err := exec.Command("file", "-bi", UPLOAD_DIR+request_id+"/"+filename).Output()

	dlonly := false
	if len(request_commands) > 1 {
		if request_commands[1] == "dl" {
			dlonly = true
		} else if request_commands[1] == "info" {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
			w.Header().Set("Cache-Control", "max-age=31536000")

			/* Filename */
			base64_t, _ := base64.StdEncoding.DecodeString(filename)
			fmt.Fprintf(w, "name: %s\n", base64_t)
			fmt.Fprintf(w, "mime: %s\n", strings.Split(string(out), ";")[0])
			fmt.Fprintf(w, "size: %s\n", FormatSize(fileinfo_t.Size()))
			fmt.Fprintf(w, "uploaded: %s\n", fileinfo_t.ModTime().Format("Mon, 2 Jan 2006 15:04:05 MST"))

			/* File MD5 */
			filedat_t, _ := ioutil.ReadFile(UPLOAD_DIR + request_id + "/" + filename)
			md5_t := md5.New()
			md5_t.Write(filedat_t)
			fmt.Fprintf(w, "md5: %x\n", md5_t.Sum(nil))

			sha1_t := sha1.New()
			sha1_t.Write(filedat_t)
			fmt.Fprintf(w, "sha1: %x\n", sha1_t.Sum(nil))
			return
		} else if request_commands[1] == "delete" && len(request_commands[1]) > 0 {

			filedat_t, _ := ioutil.ReadFile(UPLOAD_DIR + request_id + "/" + filename)
			md5_t := md5.New()
			md5_t.Write(filedat_t)
			md5_s := fmt.Sprintf("%x", md5_t.Sum(nil))

			delete_id := request_commands[1]
			exists, _ = Exists(DELETE_DIR + delete_id + "/" + request_id)
			if exists {
				os.RemoveAll(DELETE_DIR + delete_id)
				os.RemoveAll(FORCEDL_DIR + request_id)
				os.RemoveAll(HASH_DIR + md5_s)
				os.RemoveAll(UPLOAD_DIR + request_id)
			}
			http.Redirect(w, r, CONF_DOMAIN, 303)
			return
		}
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Description", "File Transfer")
	w.Header().Set("X-XSS-Protection", "1; mode=block")

	exists, _ = Exists(FORCEDL_DIR + request_id)
	base64_t, _ := base64.StdEncoding.DecodeString(filename)

	if StringInArray(strings.Split(string(out), ";")[0], whitelist) && !dlonly && !exists && err == nil {
		w.Header().Set("Content-Disposition", "inline; filename=\""+string(base64_t)+"\"")
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename=\""+string(base64_t)+"\"")
	}

	w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
	w.Header().Set("Cache-Control", "max-age=31536000, must-revalidate")
	w.Header().Set("Last-Modified", fileinfo_t.ModTime().Format("Mon, 2 Jan 2006 15:04:05 MST"))
	w.Header().Set("Content-Length", string(fileinfo_t.Size()))
	//w.Header().Set("X-Accel-Redirect", "/protected/"+request_id+"/"+filename)
	http.ServeFile(w, r, UPLOAD_DIR+request_id+"/"+filename)
	log.Printf("[LOG] File %s dowloaded\n", request_id)
	return
}