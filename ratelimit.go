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

func RateLimit(cfg Config, ip string) bool {
    if cfg.Toggles.UseRatelimit {
        md5_t := md5.New()
        io.WriteString(md5_t, ip)
        hash_t := hex.EncodeToString(md5_t.Sum(nil))
        exists, _ := Exists(cfg.Directories.Ratelimit + hash_t)
        if exists {
            file_t, _ := os.Open(cfg.Directories.Ratelimit + hash_t)
            fileinfo_t, _ := file_t.Stat()
            filemtime := fileinfo_t.ModTime().Unix()
            defer file_t.Close()
            if filemtime + int64(cfg.Limits.RatelimitTime) > time.Now().Unix() {
                curr_t, _ := ioutil.ReadFile(cfg.Directories.Ratelimit + hash_t)
                if uint64(curr_t[0]) == cfg.Limits.RatelimitFiles {
                    log.Printf("[LOG] %s is rate limited for exceed %d files/% seconds\n", cfg.Limits.RatelimitFiles, cfg.Limits.RatelimitTime)
                    return true
                } else {
                    WriteFileSafe(cfg.Directories.Ratelimit + hash_t, []byte{curr_t[0] + 1})
                }
                return false
            }
        } else {
            log.Printf("[LOG] Rate Limiting is currently disabled!\n")
        }
        WriteFileSafe(cfg.Directories.Ratelimit + hash_t, []byte{1})
    }
    return false
}
