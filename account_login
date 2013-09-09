package main
/*
func LogoutHandler(w http.ResponseWriter, r *http.Request, mc *memcache.Client) {
	cookie_t, _ := r.Cookie("wfsession")
	mc.Delete(cookie_t.Value)
	cookie := http.Cookie{Name: "wfsession", Value: "", Path: "/", Domain: "watfile.com", Expires: time.Now().Add(-5 * time.Minute), Secure: false, HttpOnly: true}
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "http://watfile.com", 302)
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
}*/