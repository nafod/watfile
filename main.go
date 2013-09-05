package main

import (
	"bytes"
	//"code.google.com/p/go.crypto/bcrypt"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/bradfitz/gomemcache/memcache"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	//"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	
	/* Format "IP:Port number" */
	CONF_IP     = ":31114"

	/* Used for redirects */
	CONF_DOMAIN = "http://localhost:31114/"

	CONF_MAX_FILESIZE = 10 << 20

	/* Base watfile data directories */
	DATA_DIR    = "./data-watfile"
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

func Login(u string, mc *memcache.Client) (map[string]string, string, error) {
	session := make(map[string]string)
	var contents_t []byte
	var err error
	session_elements_t := []string{"userid", "banned", "state", "avatar", "created"}
	for k := range session_elements_t {
		contents_t, err = ioutil.ReadFile(ACCOUNT_DIR + u + "/" + session_elements_t[k])
		if err != nil {
			fmt.Printf("%s\n", err)
			return nil, "", err
		}
		session[session_elements_t[k]] = string(bytes.TrimSpace(contents_t))
	}
	session["logged_in"] = "1"
	session["last_activity"] = "now"
	session["user"] = u
	key_t := UniqueID(100, false)
	item_t := memcache.Item{Key: key_t, Value: []byte("hello"), Expiration: 0}
	mc.Set(&item_t)
	return session, key_t, nil
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
	real_ip_t := r.Header.Get("X-Real-Ip")
	if real_ip_t == "" {
		real_ip_t = r.RemoteAddr
	}

	if RateLimit(real_ip_t) {
		fmt.Fprintf(w, MakeResult(r, "rate", ""))
		return
	}
	r.ParseMultipartForm(CONF_MAX_FILESIZE) // 10MB
	if r.MultipartForm != nil {
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
				if size_t > CONF_MAX_FILESIZE + 1 {
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
					old_id := GetHash(hash_t)
					final_id = UniqueID(8, true)
					os.Mkdir(UPLOAD_DIR+final_id, os.ModeDir)

                    files_t, _ := ioutil.ReadDir(UPLOAD_DIR + old_id + "/")

                    filename := ""
                    for a := range files_t {
                        if files_t[a].Name() != "." && files_t[a].Name() != ".." {
                            filename = files_t[a].Name()
                            break
                        }
                    }

                    fmt.Println(os.Symlink(UPLOAD_DIR+old_id+"/"+filename, UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0]))))
					os.Mkdir(DELETE_DIR+delete_id, os.ModeDir)
					WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
				} else {
					final_id = UniqueID(8, true)
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

		if len(final_dat) > CONF_MAX_FILESIZE + 1 {
			fmt.Fprintf(w, MakeResult(r, "error", ""))
			return
		}

		md5_t := md5.New()
		md5_t.Write(final_dat)
		hash_t := hex.EncodeToString(md5_t.Sum(nil))
		delete_id = UniqueID(30, false)
		exists_t, _ := Exists(HASH_DIR + hash_t + "/")
		if exists_t {
            old_id := GetHash(hash_t)
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
                os.Mkdir(UPLOAD_DIR+final_id, os.ModeDir)
                fmt.Println(os.Symlink(UPLOAD_DIR+old_id+"/"+filename, UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0]))))
                os.Mkdir(DELETE_DIR+delete_id, os.ModeDir)
                WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
            } else {
                final_id = old_id
            }
		} else {
			final_id = UniqueID(8, true)
			os.Mkdir(UPLOAD_DIR+final_id, os.ModeDir)
			if WriteFileSafe(UPLOAD_DIR+final_id+"/"+base64.StdEncoding.EncodeToString([]byte(r.Header["Up-Filename"][0])), final_dat) == false {
				fmt.Fprintf(w, MakeResult(r, "error", ""))
				return
			}
			os.Mkdir(HASH_DIR+hash_t, os.ModeDir)
			WriteEmptyFile(HASH_DIR + hash_t + "/" + final_id)
			os.Mkdir(DELETE_DIR+delete_id, os.ModeDir)
			WriteEmptyFile(DELETE_DIR + delete_id + "/" + final_id)
		}
	}
	log.Printf("[LOG] File uploaded, assigned ID %s with deletion ID %s\n", final_id, delete_id)
	fmt.Fprintf(w, MakeResult(r, final_id, delete_id))
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {

	whitelist := []string{"image/gif", "image/png", "image/jpeg", "image/bmp", "application/pdf", "text/plain"}

	/* Security checks */
	request_id_t := strings.TrimSpace(r.FormValue("id"))
	if len(request_id_t) == 0 {
		http.Redirect(w, r, CONF_DOMAIN, 303)
		return
	}

	request_commands := strings.Split(request_id_t, "/")
	request_id := strings.Split(request_commands[0], ".")[0]
	if len(request_id) == 0 {
		http.Redirect(w, r, CONF_DOMAIN, 303)
		return
	}

	exists, _ := Exists(UPLOAD_DIR + request_id + "/")
	if exists == false {
		http.Redirect(w, r, CONF_DOMAIN, 303)
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
		http.Redirect(w, r, CONF_DOMAIN, 303)
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
			http.Redirect(w, r, CONF_DOMAIN, 303)
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
	//w.Header().Set("X-Accel-Redirect", "/protected/"+request_id+"/"+filename)
	http.ServeFile(w, r, UPLOAD_DIR+request_id+"/"+filename)
	log.Printf("[LOG] File %s dowloaded\n", request_id)
	return
}

func APIDownloadHandler(w http.ResponseWriter, r *http.Request) {

	//request_id
	//request_command (info, delete)
	request_id := ""

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
	//base64_t, _ := base64.StdEncoding.DecodeString(filename)

	w.Header().Set("Expires", "Sun, 17 Jan 2038 19:14:07 GMT")
	w.Header().Set("Cache-Control", "max-age=31536000, must-revalidate")
	http.ServeFile(w, r, UPLOAD_DIR+request_id+"/"+filename)
	log.Printf("[LOG] File %s dowloaded via API\n", request_id)
	return
}

/*
func LogoutHandler(w http.ResponseWriter, r *http.Request, mc *memcache.Client) {
	cookie_t, _ := r.Cookie("wfsession")
	mc.Delete(cookie_t.Value)
	cookie := http.Cookie{Name: "wfsession", Value: "", Path: "/", Domain: "watfile.com", Expires: time.Now().Add(-5 * time.Minute), Secure: false, HttpOnly: true}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "http://watfile.com", 302)
}

func LoginHandler(w http.ResponseWriter, r *http.Request, mc *memcache.Client) {
	username := r.FormValue("user")
	password := r.FormValue("pass")
	match_t, _ := regexp.MatchString("[a-zA-Z0-9_]{1,15}", username)
	if match_t != true {
		//fmt.Fprintf(w, "Username is invalid\n")
		return
	}

	exists_t, _ := Exists(ACCOUNT_DIR + username)
	if exists_t != true {
		//fmt.Fprintf(w, "No such account\n")
		return
	}

	password_valid_t, _ := ioutil.ReadFile(ACCOUNT_DIR + username + "/password")
	password_valid := bytes.TrimSpace(password_valid_t)

	if bcrypt.CompareHashAndPassword(password_valid, []byte(password)) != nil {
		//fmt.Fprintf(w, "Invalid password\n")
		return
	}
	session, sessid, _ := Login(username, mc)
	ret, _ := mc.Get(sessid)
	cookie := http.Cookie{Name: "wfsession", Value: sessid, Path: "/", Domain: "watfile.com", Expires: time.Now().Add(5 * time.Minute), Secure: false, HttpOnly: true}
	http.SetCookie(w, &cookie)
	fmt.Fprintf(w, "Session (%s | %+v): %+v\n", sessid, ret, session)
}

func RegisterHandler(w http.ResponseWriter, r *http.Request, mc *memcache.Client) {
	username_t := r.FormValue("user")
	password_t := r.FormValue("pass")
	fmt.Fprintf(w, "Username: %s\nPassword: %s\n", username_t, password_t)
	match_t, _ := regexp.MatchString("[a-zA-Z0-9_]{1,15}", username_t)
	if match_t != true {
		fmt.Fprintf(w, "Username is invalid\n")
		return
	}

	username := username_t

	exists_t, _ := Exists(ACCOUNT_DIR + username)
	if exists_t != false {
		fmt.Fprintf(w, "Account already exists!\n")
		return
	}

	password, _ := bcrypt.GenerateFromPassword([]byte(password_t), 12)
	os.Mkdir(ACCOUNT_DIR+username, os.ModeDir)
	WriteFileSafe(ACCOUNT_DIR+username+"/username", []byte(username))
	WriteFileSafe(ACCOUNT_DIR+username+"/banned", []byte{48})
	WriteFileSafe(ACCOUNT_DIR+username+"/password", password)
	WriteFileSafe(ACCOUNT_DIR+username+"/state", []byte{48})
	WriteFileSafe(ACCOUNT_DIR+username+"/avatar", []byte("avatar"))
	WriteFileSafe(ACCOUNT_DIR+username+"/views", []byte{48})
	WriteFileSafe(ACCOUNT_DIR+username+"/userid", []byte(UniqueID(10, false)))
	WriteFileSafe(ACCOUNT_DIR+username+"/created", []byte(strconv.FormatInt(time.Now().Unix(), 10)))
	WriteEmptyFile(ACCOUNT_DIR + username + "/comments")
	WriteEmptyFile(ACCOUNT_DIR + username + "/list")

	session, sessid, _ := Login(username, mc)
	fmt.Fprintf(w, "Session (%s): %+v\n", sessid, session)
}
*/
func BlitzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "42")
}

/* Check for existence of data directories */
func Init() {
	exists, _ := Exists(DATA_DIR)
	if exists == false {
		os.Mkdir(DATA_DIR, os.ModeDir)
		log.Printf("[LOG] Initializing data directory")
	}
	exists, _ = Exists(UPLOAD_DIR)
	if exists == false {
		os.Mkdir(UPLOAD_DIR, os.ModeDir)
		log.Printf("[LOG] Initializing upload directory")
	}
	exists, _ = Exists(HASH_DIR)
	if exists == false {
		os.Mkdir(HASH_DIR, os.ModeDir)
		log.Printf("[LOG] Initializing file hash directory")
	}
	exists, _ = Exists(ACCOUNT_DIR)
	if exists == false {
		os.Mkdir(ACCOUNT_DIR, os.ModeDir)
		log.Printf("[LOG] Initializing user account directory")
	}
	exists, _ = Exists(DELETE_DIR)
	if exists == false {
		os.Mkdir(DELETE_DIR, os.ModeDir)
		log.Printf("[LOG] Initializing file delete metadata directory")
	}
	exists, _ = Exists(FORCEDL_DIR)
	if exists == false {
		os.Mkdir(FORCEDL_DIR, os.ModeDir)
		log.Printf("[LOG] Initializing force download metadata directory")
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UTC().UnixNano())
	//mc := memcache.New("127.0.0.1:11211")

	Init();

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		UploadHandler(w, r)
	})

	http.HandleFunc("/dl", func(w http.ResponseWriter, r *http.Request) {
		DownloadHandler(w, r)
	})

	/* API paths */
	//http.HandleFunc("/api/v1/upload", func(w http.ResponseWriter, r *http.Request) {
	//	APIUploadHandler(w, r)
	//})

	http.HandleFunc("/api/v1/dl", func(w http.ResponseWriter, r *http.Request) {
		APIDownloadHandler(w, r)
	})

	/*http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		LoginHandler(w, r, mc)
	})

	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		LogoutHandler(w, r, mc)
	})

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		RegisterHandler(w, r, mc)
	})*/

	http.HandleFunc("/mu-3f8488db-7fabdac2-b1583628-30caf91d", BlitzHandler)
	log.Fatal(http.ListenAndServe(CONF_IP, nil))
}
