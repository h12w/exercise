package main

import (
	"log"
	"net/http"
	"strings"

	"h12.io/httpauth"
)

func getLogin(rw http.ResponseWriter, req *http.Request) {
	message := strings.Join(auth.Messages(rw, req), " ")
	//rw.Header().Add("login-messsage", message)
	if err := tem.ExecuteTemplate(rw, "login.html", message); err != nil {
		log.Println(err.Error())
	}
}

func postLogin(rw http.ResponseWriter, req *http.Request) {
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	if err := auth.Login(rw, req, username, password, "/"); err != nil && err.Error() == "already authenticated" {
		http.Redirect(rw, req, "/", http.StatusSeeOther)
	} else if err != nil {
		log.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	}
}

func handleLogout(rw http.ResponseWriter, req *http.Request) {
	user, err := auth.Authorize(rw, req)
	if err != nil {
		log.Println("fail to authorize")
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
		return
	}
	log.Println("logout")
	if err := auth.Logout(rw, req); err != nil {
		log.Println(err)
	}
	if players.Remove(user) {
		if newUser := waiters.PopFront(); newUser != nil {
			players.Add(newUser)
			log.Println(user.Name, "is remove, added", newUser.Name)
		}
	} else {
		waiters.Remove(user)
	}
	http.Redirect(rw, req, "/", http.StatusSeeOther)
}

func postRegister(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Name = req.PostFormValue("username")
	password := req.PostFormValue("password")
	auth.Register(rw, req, user, password)
	http.Redirect(rw, req, "/login", http.StatusSeeOther)
}
