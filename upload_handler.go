package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
    "time"
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

func UploadHandler(cfg Config, w http.ResponseWriter, r *http.Request, db *sql.DB) {
	delete_id := ""
	final_id := ""

	/* Determine user IP */
	real_ip_t := r.Header.Get("X-Real-Ip")
	if real_ip_t == "" {
		real_ip_t = r.RemoteAddr
	}

	/* Check if IP is currently ratelimited */
	if RateLimit(cfg, real_ip_t) {
		fmt.Fprintf(w, MakeResult(r, "rate", ""))
		return
	}

	r.ParseMultipartForm(int64(cfg.Limits.MaxFilesize))
	var ret_files []UploadedFile

	files_t, ok := r.MultipartForm.File["upload"]

	/* No files actually uploaded */
	if ok != true || len(files_t) == 0 {
		fmt.Fprintf(w, MakeResult(r, "error", ""))
		return
	}

	ret_files = make([]UploadedFile, len(r.MultipartForm.File["upload"]))

	dbtxt, err := db.Begin()
	if err != nil {
		/* Database problem, likely fatal */
        dbtxt.Rollback()
		panic(err)
	}

	for _, file_t := range files_t {
		buffer_t := make([]byte, cfg.Limits.MaxFilesize+1)
		f, err := file_t.Open()
		defer f.Close()

		/* Couldn't read the file */
		size_t, err := f.Read(buffer_t)
		if err != nil {
			ret_files = append(ret_files, UploadedFile{"", "", "error"})
			continue
		}

		/* File is bigger than the maximum filesize */
		if uint64(size_t) > cfg.Limits.MaxFilesize+1 {
			ret_files = append(ret_files, UploadedFile{"", "", "error"})
			continue
		}

		/* Strip fial character from buffer */
		buffer_t = buffer_t[:size_t]

		/* Generate MD5 hash of file */
		md5_t := md5.New()
		md5_t.Write(buffer_t)
		hash_t := hex.EncodeToString(md5_t.Sum(nil))

		/* Unique file ID and deletion ID */
		final_id = UniqueID(cfg, 8, true)
		delete_id = UniqueID(cfg, 30, false)

		/* Check if file has already been uploaded (de-duplication) */
		exists_t, err := Exists(cfg.Directories.Upload + hash_t)
		if err != nil {
			ret_files = append(ret_files, UploadedFile{"", "", "error"})
			continue
		}

        dbstmt, err := dbtxt.Prepare("INSERT INTO files(name, size, md5, fileid, diskid, deleteid, uploaded, downloads, views, author) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
        if err != nil {
            dbtxt.Rollback()
            panic(err)
        }
        defer dbstmt.Close()

        _, err = dbstmt.Exec(file_t.Filename, size_t, hash_t, final_id, hash_t, delete_id, time.Now().Unix(), 0, 0, 0)
        if err != nil {
            dbtxt.Rollback()
            panic(err)
        }
        if exists_t == false {
            ioutil.WriteFile(cfg.Directories.Upload + hash_t, buffer_t, os.ModePerm)
        }
		ret_files = append(ret_files, UploadedFile{file_t.Filename, final_id, ""})
	}

	log.Printf("[LOG] File uploaded, assigned ID %s with deletion ID %s\n", final_id, delete_id)

	/* Put changes in database */
	dbtxt.Commit()
	/* Create response */
	json_out, err := json.Marshal(map[string][]UploadedFile{"files": ret_files[1:]})
	if err != nil {
		/* Panic - unable to create JSON */
		panic(err)
	}

	fmt.Fprintf(w, string(json_out))
}
