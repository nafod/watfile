package main

import (
	"fmt"
	"os"
	"strconv"
	"math/rand"
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

func UniqueID(todo int, exists bool) string {
	ret_t := strconv.FormatUint(uint64(rand.Int63n(4294967295)), 36)
	exists_t := exists
	for exists_t {
		ret_t = strconv.FormatUint(uint64(rand.Int63n(4294967295)), 36)
		exists_t, _ = Exists(UPLOAD_DIR + ret_t)
	}
	return ret_t
}