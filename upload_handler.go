package main

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	delete_id := ""
	final_id := ""
	real_ip_t := r.Header.Get("X-Real-Ip")
    perm := os.ModeDir | 0744

	if real_ip_t == "" {
		real_ip_t = r.RemoteAddr
	}

	if RateLimit(real_ip_t) {
		fmt.Fprintf(w, MakeResult(r, "rate", ""))
		return
	}

	r.ParseMultipartForm(CONF_MAX_FILESIZE)
    log.Printf("Received request: %+v\n", r.MultipartForm)
	if r.MultipartForm != nil {
		if _, ok := r.MultipartForm.File["upload"]; ok {
			if len(r.MultipartForm.File["upload"]) == 0 {
				fmt.Fprintf(w, MakeResult(r, "error", ""))
				return
			}
			files_t := r.MultipartForm.File["upload"]
			for _, file_t := range files_t {
				buffer_t := make([]byte, CONF_MAX_FILESIZE+1)
				f, err := file_t.Open()
				defer f.Close()

				size_t, err := f.Read(buffer_t)
				if err != nil {
					fmt.Fprintf(w, MakeResult(r, "error", ""))
					return
				}
				if size_t > CONF_MAX_FILESIZE+1 {
					fmt.Fprintf(w, MakeResult(r, "size", ""))
					return
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

					fmt.Println(os.Symlink(UPLOAD_DIR+old_id+"/"+filename, UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0]))))
					os.Mkdir(DELETE_DIR+delete_id, perm)
					WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
				} else {
					final_id = UniqueID(8, true)
					os.Mkdir(UPLOAD_DIR+final_id, perm)
					if WriteFileSafe(UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0])), buffer_t) == false {
						fmt.Fprintf(w, MakeResult(r, "error", ""))
						return
					}
					os.Mkdir(HASH_DIR+hash_t, perm)
					WriteEmptyFile(HASH_DIR + hash_t + "/" + final_id)
					os.Mkdir(DELETE_DIR+delete_id, perm)
					WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
					/* Check image size and create forcedl here */
				}

			}
		} else {
			fmt.Fprintf(w, MakeResult(r, "error", ""))
			return
		}
	} else {
		p := new(bytes.Buffer)
		p.ReadFrom(r.Body)
		final_dat := p.Bytes()
		if r.FormValue("base64") == "true" {
			base64_t, _ := base64.StdEncoding.DecodeString(p.String())
			final_dat = []byte(base64_t)
		}

		if len(final_dat) > CONF_MAX_FILESIZE+1 {
			fmt.Fprintf(w, MakeResult(r, "error", ""))
			return
		}

		md5_t := md5.New()
		md5_t.Write(final_dat)
		hash_t := hex.EncodeToString(md5_t.Sum(nil))
		delete_id = UniqueID(30, false)
		exists_t, _ := Exists(HASH_DIR + hash_t + "/")
		if exists_t {
			old_id := GetIDHash(hash_t)
			files_t, _ := ioutil.ReadDir(UPLOAD_DIR + old_id + "/")

			filename := ""
			for a := range files_t {
				if files_t[a].Name() != "." && files_t[a].Name() != ".." {
					filename = files_t[a].Name()
					break
				}
			}

			filename_new := base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0]))
			if filename_new != filename {
				final_id = UniqueID(8, true)
				os.Mkdir(UPLOAD_DIR+final_id, perm)
				fmt.Println(os.Symlink(UPLOAD_DIR+old_id+"/"+filename, UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0]))))
				os.Mkdir(DELETE_DIR+delete_id, perm)
				WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
			} else {
				final_id = old_id
			}
		} else {
			final_id = UniqueID(8, true)
			os.Mkdir(UPLOAD_DIR+final_id, perm)
			if WriteFileSafe(UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0])), final_dat) == false {
				fmt.Fprintf(w, MakeResult(r, "error", ""))
				return
			}
			os.Mkdir(HASH_DIR+hash_t, perm)
			WriteEmptyFile(HASH_DIR + hash_t + "/" + final_id)
			os.Mkdir(DELETE_DIR+delete_id, perm)
			WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
		}
	}
	log.Printf("[LOG] File uploaded, assigned ID %s with deletion ID %s\n", final_id, delete_id)
	fmt.Fprintf(w, MakeResult(r, final_id, delete_id))
}
