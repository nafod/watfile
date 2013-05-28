package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	CONF_IP     = ":31114"
	DATA_DIR    = "/var/www/data-watfile/test"
	UPLOAD_DIR  = DATA_DIR + "/uploads/"
	HASH_DIR    = DATA_DIR + "/hashes/"
	ACCOUNT_DIR = DATA_DIR + "/accounts/"
	DELETE_DIR  = DATA_DIR + "/delete/"
	FORCEDL_DIR = DATA_DIR + "/forcedl/"
)

func StringInArray(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func FormatSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	id := 0
	size_t := size
	for id < len(units) && size_t > 1024 {
		size_t = size_t / 1024
		id = id + 1
	}
	return fmt.Sprintf("%d %s", size_t, units[id])
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func UniqueID(todo int) string {
	const alphanum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	p := make([]byte, todo)
	offset := 0
	for {
		val := int64(rand.Int())
		for i := 0; i < 8; i++ {
			p[offset] = alphanum[int(val&0xff)%len(alphanum)]
			todo--
			if todo == 0 {
				return string(p)
			}
			offset++
			val >>= 8
		}
	}

	panic("unreachable")
}

func MakeResult(req *http.Request, t string, del string) string {
	fmt.Printf("Gottem: |%s| %d\n", t, len(t))
	if val, ok := req.Header["Up-Id"]; ok {
		return string(val[0]) + "|" + del + "|" + t
	}
	return "0|" + del + "|" + t
}

func WriteFileSafe(path string, content []byte) bool {
	ioutil.WriteFile(path, content, os.ModePerm)
	return true
}

func WriteEmptyFile(path string) bool {
	ioutil.WriteFile(path, []byte{}, os.ModePerm)
	return true
}

func GetHash(hash string) string {
	files_t, _ := ioutil.ReadDir(HASH_DIR + hash + "/")
	for a := range files_t {
		if files_t[a].Name() != "." && files_t[a].Name() != ".." {
			return files_t[a].Name()
		}
	}
	return ""
}

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	delete_id := ""
	final_id := ""
	r.ParseMultipartForm(10 << 20) // 10MB
	if _, ok := r.MultipartForm.File["upload"]; ok {
		if len(r.MultipartForm.File["upload"]) == 0 {
			fmt.Fprintf(w, MakeResult(r, "error", ""))
			return
		}
		files_t := r.MultipartForm.File["upload"]
		for _, file_t := range files_t {
			buffer_t := make([]byte, 10<<20+1)
			f, err := file_t.Open()
			defer f.Close()

			size_t, err := f.Read(buffer_t)
			if err != nil {
				fmt.Fprintf(w, MakeResult(r, "error", ""))
				return
			}
			if size_t > 10485761 {
				fmt.Fprintf(w, MakeResult(r, "size", ""))
				return
			}

			fmt.Printf("Filesize: %d\n", size_t)

			buffer_t = buffer_t[:size_t]

			md5_t := md5.New()
			md5_t.Write(buffer_t)
			hash_t := hex.EncodeToString(md5_t.Sum(nil))
			fmt.Printf("hash is: %+v\n", hash_t)
			delete_id = UniqueID(30)
			exists_t, _ := Exists(HASH_DIR + hash_t + "/")
			if exists_t {
				final_id = GetHash(hash_t)
				fmt.Printf("|%s|\n", final_id)
			} else {
				final_id = UniqueID(8)
				fmt.Printf("LOG: %s\n", final_id)
				os.Mkdir(UPLOAD_DIR+final_id, os.ModeDir)
				if WriteFileSafe(UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0])), buffer_t) == false {
					fmt.Fprintf(w, MakeResult(r, "error", ""))
					return
				}
				os.Mkdir(HASH_DIR+hash_t, os.ModeDir)
				WriteEmptyFile(HASH_DIR + hash_t + "/" + final_id)
				os.Mkdir(DELETE_DIR+delete_id, os.ModeDir)
				WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
				/* Check image size and create forcedl here */
			}

		}
	} else if r.FormValue("up") == "true" {
		p := new(bytes.Buffer)
		p.ReadFrom(r.Body)
		final_dat := p.Bytes()
		if r.FormValue("base64") == "true" {
			base64_t, _ := base64.StdEncoding.DecodeString(p.String())
			final_dat = []byte(base64_t)
		}

		if len(final_dat) > 10485761 {
			fmt.Fprintf(w, MakeResult(r, "error", ""))
			return
		}

		md5_t := md5.New()
		md5_t.Write(final_dat)
		hash_t := hex.EncodeToString(md5_t.Sum(nil))
		fmt.Printf("hash is: %+v\n", hash_t)
		delete_id = UniqueID(30)
		exists_t, _ := Exists(HASH_DIR + hash_t + "/")
		if exists_t {
			final_id = GetHash(hash_t)
			fmt.Printf("|%s|\n", final_id)
		} else {
			final_id = UniqueID(8)
			fmt.Printf("LOG: %s\n", final_id)
			os.Mkdir(UPLOAD_DIR+final_id, os.ModeDir)
			if WriteFileSafe(UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0])), final_dat) == false {
				fmt.Fprintf(w, MakeResult(r, "error", ""))
				return
			}
			os.Mkdir(HASH_DIR+hash_t, os.ModeDir)
			WriteEmptyFile(HASH_DIR + hash_t + "/" + final_id)
			os.Mkdir(DELETE_DIR+delete_id, os.ModeDir)
			WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
			/* Check image size and create forcedl here */
		}

		fmt.Printf("%s\n", r.Header["Up-Filename"][0])
	} else {
		fmt.Fprintf(w, MakeResult(r, "error", ""))
	}
	fmt.Fprintf(w, MakeResult(r, final_id, delete_id))
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {

	whitelist := []string{"image/gif", "image/png", "image/jpeg", "image/bmp", "application/pdf", "text/plain"}

	/* Security checks */
	request_id_t := strings.TrimSpace(r.FormValue("id"))
	if len(request_id_t) == 0 {
		http.Redirect(w, r, "http://watfile.com/", 303)
		return
	}

	request_commands := strings.Split(request_id_t, "/")
	request_id := strings.Split(request_commands[0], ".")[0]
	if len(request_id) == 0 {
		http.Redirect(w, r, "http://watfile.com/", 303)
		return
	}

	fmt.Printf("%s\n", request_id)
	exists, _ := Exists(UPLOAD_DIR + request_id + "/")
	if exists == false {
		http.Redirect(w, r, "http://watfile.com/", 303)
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
		http.Redirect(w, r, "http://watfile.com/", 303)
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
			http.Redirect(w, r, "http://watfile.com/", 303)
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
	http.ServeFile(w, r, UPLOAD_DIR+request_id+"/"+filename)
	return
}

func BlitzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "42")
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "/var/www/watfile/index.html")
	})
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		UploadHandler(w, r)
	})
	http.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) {
		DownloadHandler(w, r)
	})
	http.HandleFunc("/mu-3f8488db-7fabdac2-b1583628-30caf91d", BlitzHandler)
	log.Fatal(http.ListenAndServe(CONF_IP, nil))
}
