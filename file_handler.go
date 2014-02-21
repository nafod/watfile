package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func FileHandler(cfg Config, w http.ResponseWriter, r *http.Request) {

	//whitelist := []string{"image/gif", "image/png", "image/jpeg", "image/bmp", "application/pdf", "text/plain"}

	/* Security checks */
	request_id_t := strings.TrimSpace(r.FormValue("id"))
	if len(request_id_t) == 0 {
		http.Redirect(w, r, cfg.Main.Domain, 303)
		return
	}

	request_commands := strings.Split(request_id_t, "/")
	request_id := strings.Split(request_commands[0], ".")[0]
	if len(request_id) == 0 {
		http.Redirect(w, r, cfg.Main.Domain, 303)
		return
	}

	exists, _ := Exists(cfg.Directories.Upload + request_id + "/")
	if exists == false {
		http.Redirect(w, r, cfg.Main.Domain, 303)
		return
	}

	files_t, _ := ioutil.ReadDir(cfg.Directories.Upload + request_id + "/")

	filename := ""
	for a := range files_t {
		if files_t[a].Name() != "." && files_t[a].Name() != ".." {
			filename = files_t[a].Name()
			break
		}
	}

	if len(filename) == 0 {
		http.Redirect(w, r, cfg.Main.Domain, 303)
		return
	}

	fileinfo_t, _ := os.Stat(cfg.Directories.Upload + request_id + "/" + filename)
	out, err := exec.Command("file", "-biL", cfg.Directories.Upload+request_id+"/"+filename).Output()
	if err != nil {
		log.Printf("[ERROR] Unable to determine mine of file %s\n", request_id)
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
	w.Header().Set("Cache-Control", "max-age=31536000")

	/* Filename */
	base64_t, _ := base64.URLEncoding.DecodeString(filename)
	fmt.Fprintf(w, "name: %s\n", base64_t)
	fmt.Fprintf(w, "mime: %s\n", strings.Split(string(out), ";")[0])
	fmt.Fprintf(w, "size: %s\n", FormatSize(fileinfo_t.Size()))
	fmt.Fprintf(w, "uploaded: %s\n", fileinfo_t.ModTime().Format("Mon, 2 Jan 2006 15:04:05 MST"))

	/* File MD5 */
	filedat_t, _ := ioutil.ReadFile(cfg.Directories.Upload + request_id + "/" + filename)
	md5_t := md5.New()
	md5_t.Write(filedat_t)
	fmt.Fprintf(w, "md5: %x\n", md5_t.Sum(nil))

	sha1_t := sha1.New()
	sha1_t.Write(filedat_t)
	fmt.Fprintf(w, "sha1: %x\n", sha1_t.Sum(nil))
	fmt.Fprintf(w, "Download file: %sdl?id=%s\n", cfg.Main.Domain, request_id)
	log.Printf("[LOG] File %s viewed\n", request_id)
	return
}
