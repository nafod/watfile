package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func DeleteHandler(cfg Config, w http.ResponseWriter, r *http.Request) {

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

	filedat_t, _ := ioutil.ReadFile(cfg.Directories.Upload + request_id + "/" + filename)
	md5_t := md5.New()
	md5_t.Write(filedat_t)
	md5_s := fmt.Sprintf("%x", md5_t.Sum(nil))

	delete_id := request_commands[1]
	exists, _ = Exists(cfg.Directories.Delete + delete_id + "/" + request_id)
	if exists {
		os.RemoveAll(cfg.Directories.Delete + delete_id)
		os.RemoveAll(cfg.Directories.ForceDL + request_id)
		os.RemoveAll(cfg.Directories.Hash + md5_s)
		os.RemoveAll(cfg.Directories.Upload + request_id)
	}
	http.Redirect(w, r, cfg.Main.Domain, 303)
	log.Printf("[LOG] File %s deleted\n", request_id)
	return
}
