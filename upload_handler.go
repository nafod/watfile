package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type UploadedFile struct {
	Name  string
	ID    string
	Error string
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	delete_id := ""
	final_id := ""
	real_ip_t := r.Header.Get("X-Real-Ip")
	perm := os.ModeDir | 0744
	log.Printf("Received request: %+v\n", r)

	if real_ip_t == "" {
		real_ip_t = r.RemoteAddr
	}

	if RateLimit(real_ip_t) {
		fmt.Fprintf(w, MakeResult(r, "rate", ""))
		return
	}

	r.ParseMultipartForm(CONF_MAX_FILESIZE)
    var ret_files []UploadedFile
	if _, ok := r.MultipartForm.File["upload"]; ok {
		if len(r.MultipartForm.File["upload"]) == 0 {
			fmt.Fprintf(w, MakeResult(r, "error", ""))
			return
		}
		files_t := r.MultipartForm.File["upload"]
        ret_files = make([]UploadedFile, len(r.MultipartForm.File["upload"]))
		for _, file_t := range files_t {
			buffer_t := make([]byte, CONF_MAX_FILESIZE+1)
			f, err := file_t.Open()
			defer f.Close()

			size_t, err := f.Read(buffer_t)
			if err != nil {
				ret_files = append(ret_files, UploadedFile{"", "", "error"})
				continue
			}
			if size_t > CONF_MAX_FILESIZE+1 {
				ret_files = append(ret_files, UploadedFile{"", "", "error"})
				continue
			}

			buffer_t = buffer_t[:size_t]

			md5_t := md5.New()
			md5_t.Write(buffer_t)
			hash_t := hex.EncodeToString(md5_t.Sum(nil))
			delete_id = UniqueID(30, false)
			exists_t, _ := Exists(HASH_DIR + hash_t + "/")
			if exists_t {
				old_id := GetIDHash(hash_t)
				final_id = UniqueID(8, true)
				os.Mkdir(UPLOAD_DIR+final_id, perm)

				files_t, _ := ioutil.ReadDir(UPLOAD_DIR + old_id + "/")

				filename := ""
				for a := range files_t {
					if files_t[a].Name() != "." && files_t[a].Name() != ".." {
						filename = files_t[a].Name()
						break
					}
				}

				fmt.Println(os.Symlink(UPLOAD_DIR+old_id+"/"+filename, UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(file_t.Filename))))
				os.Mkdir(DELETE_DIR+delete_id, perm)
				WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
			} else {
				final_id = UniqueID(8, true)
				os.Mkdir(UPLOAD_DIR+final_id, perm)
				if WriteFileSafe(UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(file_t.Filename)), buffer_t) == false {
                    ret_files = append(ret_files, UploadedFile{"", "", "error"})
					continue
				}
				os.Mkdir(HASH_DIR+hash_t, perm)
				WriteEmptyFile(HASH_DIR + hash_t + "/" + final_id)
				os.Mkdir(DELETE_DIR+delete_id, perm)
				WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
				/* Check image size and create forcedl here */
			}
            ret_files = append(ret_files, UploadedFile{file_t.Filename, final_id, ""})
		}
	} else {
		fmt.Fprintf(w, MakeResult(r, "error", ""))
		return
	}
	log.Printf("[LOG] File uploaded, assigned ID %s with deletion ID %s\n", final_id, delete_id)
	//fmt.Fprintf(w, MakeResult(r, final_id, delete_id))
	json_out, err := json.Marshal(map[string][]UploadedFile{"files": ret_files[1:]})
    if err != nil {
        panic(err)
    }
	fmt.Fprintf(w, string(json_out))
}
