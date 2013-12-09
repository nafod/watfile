package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func RateLimit(ip string) bool {
    if CONF_USE_RATELIMITING {
        md5_t := md5.New()
        io.WriteString(md5_t, ip)
        hash_t := hex.EncodeToString(md5_t.Sum(nil))
        exists, _ := Exists(RATELIMIT_DIR + hash_t)
        if exists {
            file_t, _ := os.Open(RATELIMIT_DIR + hash_t)
            fileinfo_t, _ := file_t.Stat()
            filemtime := fileinfo_t.ModTime().Unix()
            defer file_t.Close()
            if filemtime + CONF_RATELIMIT_TIME > time.Now().Unix() {
                curr_t, _ := ioutil.ReadFile(RATELIMIT_DIR + hash_t)
                if curr_t[0] == CONF_RATELIMIT_FILES {
                    log.Printf("[LOG] %s is rate limited for exceed %d files/% seconds\n", CONF_RATELIMIT_FILES, CONF_RATELIMIT_TIME)
                    return true
                } else {
                    WriteFileSafe(RATELIMIT_DIR + hash_t, []byte{curr_t[0] + 1})
                }
                return false
            }
        } else {
            log.Printf("[LOG] Rate Limiting is currently disabled!\n")
        }
        WriteFileSafe(RATELIMIT_DIR + hash_t, []byte{1})
    }
    return false
}
