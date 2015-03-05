package main

import (
	"container/list"
	"fmt"

	"sync"
	"time"

	"golang.org/x/net/websocket"
	"h12.me/httpauth"
)

func serveWaitNum(ws *websocket.Conn) {
	//io.Copy(ws, ws)
	fmt.Fprintf(ws, "5")
	time.Sleep(time.Second)
	fmt.Fprintf(ws, "2")
	//log.Println(ws.Request())
}

type UserQueue struct {
	l  list.List
	mu sync.Mutex
}

func (q *UserQueue) Push(u *httpauth.UserData) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.l.PushBack(u)
}

func (q *UserQueue) Pop() *httpauth.UserData {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.l.Len() > 0 {
		return q.l.Remove(q.l.Front()).(*httpauth.UserData)
	}
	return nil
}
