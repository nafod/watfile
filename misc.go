package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
)

func MakeResult(req *http.Request, t string, del string) string {
	if val, ok := req.Header["Up-Id"]; ok {
		if del != "" {
			return fmt.Sprintf(`{"uid": %s, "file": "%s", "del": "%s"}`, val[0], t, del)
		}
		return fmt.Sprintf(`{"uid": %s, "err": "%s"`, val[0], t)
	}
	if del != "" {
		return fmt.Sprintf(`{"uid": %s, "file": "%s", "del": "%s"}`, 0, t, del)
	}
	return fmt.Sprintf(`{"uid": %s, "err": "%s"`, 0, t)
}

func GetIDHash(cfg Config, hash string) string {
	files_t, _ := ioutil.ReadDir(cfg.Directories.Hash + hash + "/")
	for a := range files_t {
		if files_t[a].Name() != "." && files_t[a].Name() != ".." {
			return files_t[a].Name()
		}
	}
	return ""
}

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

func UniqueID(cfg Config, todo int, exists bool) string {
	ret_t := strconv.FormatUint(uint64(rand.Int63n(4294967295)), 36)
	exists_t := exists

	for exists_t {
		ret_t = strconv.FormatUint(uint64(rand.Int63n(4294967295)), 36)
		exists_t, _ = Exists(cfg.Directories.Upload + ret_t)
	}
	return ret_t
}
