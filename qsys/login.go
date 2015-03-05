package main

import (
	"log"
	"net/http"
	"strings"

	"h12.me/httpauth"
)

func getLogin(rw http.ResponseWriter, req *http.Request) {
	messages := strings.Join(auth.Messages(rw, req), " ")
	if err := tem.ExecuteTemplate(rw, "login.html", messages); err != nil {
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
	user, err := auth.Authorize(rw, req, true)
	if err != nil {
		log.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
		return
	}
	players.Remove(user)
	if err := auth.Logout(rw, req); err != nil {
		log.Println(err)
	}
	http.Redirect(rw, req, "/", http.StatusSeeOther)
}

func postRegister(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Username = req.PostFormValue("username")
	user.Email = req.PostFormValue("email")
	password := req.PostFormValue("password")
	if err := auth.Register(rw, req, user, password); err == nil {
		postLogin(rw, req)
	} else {
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	}
}
