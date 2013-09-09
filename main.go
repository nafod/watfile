package main

import (
	//"bytes"
	//"code.google.com/p/go.crypto/bcrypt"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	//"regexp"
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

/*func Login(u string, mc *memcache.Client) (map[string]string, string, error) {
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
}*/

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

func main() {

	Init()

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

	http.HandleFunc("/mu-3f8488db-7fabdac2-b1583628-30caf91d", func (w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "42")
	})

	log.Fatal(http.ListenAndServe(CONF_IP, nil))
}
