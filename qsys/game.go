package main

import (
	"log"
	"net/http"
	"sync"

	"h12.me/httpauth"
)

func handleGame(rw http.ResponseWriter, req *http.Request) {
	user, err := auth.Authorize(rw, req, true)
	if err != nil {
		log.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
		return
	}
	type data struct {
		User *httpauth.UserData
	}
	d := data{User: user}
	templ := "game.html"
	if !auth.Satisfy(user, "player") {
		templ = "wait.html"
	}
	if err := tem.ExecuteTemplate(rw, templ, d); err != nil {
		log.Println(err.Error())
	}
}

type playerPool struct {
	capacity int
	mu       sync.Mutex
}

func (p *playerPool) Remove(user *httpauth.UserData) {
	if !auth.Satisfy(user, "player") {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.capacity--
}

func (p *playerPool) Add(user *httpauth.UserData) {
	if auth.Satisfy(user, "player") {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.capacity++
}
