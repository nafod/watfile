package main

/*
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
}*/
