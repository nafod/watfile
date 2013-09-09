package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"time"
)

func RateLimit(ip string) bool {
	md5_t := md5.New()
	io.WriteString(md5_t, ip)
	hash_t := hex.EncodeToString(md5_t.Sum(nil))
	exists, _ := Exists(DATA_DIR + "/ratelimit/" + hash_t)
	if exists {
		file_t, _ := os.Open(DATA_DIR + "/ratelimit/" + hash_t)
		fileinfo_t, _ := file_t.Stat()
		filemtime := fileinfo_t.ModTime().Unix()
		defer file_t.Close()
		if filemtime+300 > time.Now().Unix() {
			curr_t, _ := ioutil.ReadFile(DATA_DIR + "/ratelimit/" + hash_t)
			if curr_t[0] == 30 {
				return true
			} else {
				WriteFileSafe(DATA_DIR+"/ratelimit/"+hash_t, []byte{curr_t[0] + 1})
			}
			return false
		}
	}
	WriteFileSafe(DATA_DIR+"/ratelimit/"+hash_t, []byte{1})
	return false
}
