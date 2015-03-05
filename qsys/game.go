package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

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
	//io.Copy(ws, ws)
	fmt.Fprintf(ws, "5")
	time.Sleep(time.Second)
	fmt.Fprintf(ws, "2")
	//log.Println(ws.Request())
}
