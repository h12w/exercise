package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"

	"h12.me/httpauth"
)

func handleGame(rw http.ResponseWriter, req *http.Request) {
	user, err := auth.Authorize(rw, req)
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
	if !players.Add(user) {
		waiters.PushBack(user)
		templ = "wait.html"
	}
	if err := tem.ExecuteTemplate(rw, templ, d); err != nil {
		log.Println(err.Error())
	}
}

func serveWaitNum(ws *websocket.Conn) {
	user, _ := auth.Authorize(nil, ws.Request())
	ch := make(chan int, 1)
	i := waiters.Register(user.Name, ch)
	fmt.Fprintf(ws, "%d", i)
	for i := range ch {
		fmt.Fprintf(ws, "%d", i)
	}
}
