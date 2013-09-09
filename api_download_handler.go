package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

func APIDownloadHandler(w http.ResponseWriter, r *http.Request) {

	//request_id
	//request_command (info, delete)
	r.ParseForm()
	log.Printf("Form values: %+v\n", r.Form)
	request_id := r.Form.Get("file")

	/* Security checks */
	if len(request_id) == 0 {
		fmt.Fprintf(w, "invalid file")
		return
	}

	exists, _ := Exists(UPLOAD_DIR + request_id + "/")
	if exists == false {
		fmt.Fprintf(w, "no such file")
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
		fmt.Fprintf(w, "invalid filename")
		return
	}

	/* Original filename */
	base64_t, _ := base64.StdEncoding.DecodeString(filename)

	w.Header().Set("Content-Disposition", "attachment; filename=\""+string(base64_t)+"\"")
	w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
	w.Header().Set("Cache-Control", "max-age=31536000, must-revalidate")
	http.ServeFile(w, r, UPLOAD_DIR+request_id+"/"+filename)
	log.Printf("[LOG] File %s dowloaded via API\n", request_id)
	return
}
