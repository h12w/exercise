package main

import (
	"container/list"
	"log"
	"sync"

	"h12.io/httpauth"
)

type User struct {
	*httpauth.UserData
	c chan *Message
}

type Message struct {
	Num int
}

type PlayerPool struct {
	m        map[string]*httpauth.UserData
	capacity int
	mu       sync.Mutex
}

func NewPlayerPool() PlayerPool {
	return PlayerPool{m: make(map[string]*httpauth.UserData)}
}

func (p *PlayerPool) Count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.m)
}

func (p *PlayerPool) Remove(user *httpauth.UserData) (removed bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, in := p.m[user.Name]; !in {
		return false
	}
	delete(p.m, user.Name)
	p.capacity++
	log.Println("capacity++", p.capacity)
	return true
}

func (p *PlayerPool) Add(user *httpauth.UserData) (isPlayer bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, in := p.m[user.Name]; in {
		return true
	}
	if p.capacity == 0 {
		return false
	}
	p.m[user.Name] = user
	p.capacity--
	log.Println("capacity--", p.capacity)
	return true
}

type UserQueue struct {
	l  list.List
	mu sync.Mutex
}

func (q *UserQueue) PushBack(u *httpauth.UserData) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for e := q.l.Front(); e != nil; e = e.Next() {
		if u.Name == e.Value.(*User).Name {
			return
		}
	}
	q.l.PushBack(&User{UserData: u})
}

func (q *UserQueue) Remove(u *httpauth.UserData) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for e := q.l.Front(); e != nil; e = e.Next() {
		if u.Name == e.Value.(*User).Name {
			close(e.Value.(*User).c)
			first := e.Next()
			q.l.Remove(e)
			q.notify(first)
			return
		}
	}
}

func (q *UserQueue) PopFront() *httpauth.UserData {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.l.Len() > 0 {
		noti := q.l.Remove(q.l.Front()).(*User)
		noti.c <- &Message{0}
		close(noti.c)
		q.notify(q.l.Front())
		return noti.UserData
	}
	return nil
}

func (q *UserQueue) notify(first *list.Element) {
	i := 0
	for e := q.l.Front(); e != nil; e = e.Next() {
		if e == first {
			break
		}
		i++
	}
	for e := first; e != nil; e = e.Next() {
		noti := e.Value.(*User)
		noti.c <- &Message{i}
		i++
	}
}

func (q *UserQueue) Register(userName string, c chan *Message) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	i := 1
	for e := q.l.Front(); e != nil; e = e.Next() {
		if userName == e.Value.(*User).Name {
			if old := e.Value.(*User).c; old != nil {
				// if a user log in repeatedly, only keep the lastest websocket to reduce resouce usage
				close(old)
			}
			e.Value.(*User).c = c
			return i
		}
		i++
	}
	return 0
}
