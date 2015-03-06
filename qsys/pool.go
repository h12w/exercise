package main

import (
	"container/list"
	"log"
	"sync"

	"h12.me/httpauth"
)

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
			log.Println("duplicate", u.Name)
			return
		}
	}
	q.l.PushBack(&User{UserData: u})
}

func (q *UserQueue) PopFront() *httpauth.UserData {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.l.Len() > 0 {
		noti := q.l.Remove(q.l.Front()).(*User)
		noti.c <- &Message{0}
		close(noti.c)
		i := 0
		for e := q.l.Front(); e != nil; e = e.Next() {
			noti := e.Value.(*User)
			log.Println(noti.Name, i)
			noti.c <- &Message{i}
			i++
		}
		return noti.UserData
	}
	return nil
}

func (q *UserQueue) Register(userName string, c chan *Message) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	i := 1
	for e := q.l.Front(); e != nil; e = e.Next() {
		if userName == e.Value.(*User).Name {
			if old := e.Value.(*User).c; old != nil {
				close(old)
			}
			e.Value.(*User).c = c
			return i
		}
		i++
	}
	return 0
}
