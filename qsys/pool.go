package main

import (
	"container/list"
	"log"
	"sync"

	"h12.me/httpauth"
)

type playerPool struct {
	capacity int
	mu       sync.Mutex
}

func (p *playerPool) Remove(user *httpauth.UserData) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !auth.Satisfy(user, "player") {
		return false
	}
	auth.ChangeRole(user, "waiter")
	p.capacity++
	log.Println("capacity++", p.capacity)
	return true
}

func (p *playerPool) Add(user *httpauth.UserData) (isPlayer bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if auth.Satisfy(user, "player") {
		log.Printf("user %s is already a player", user.Name)
		return true
	}
	if p.capacity == 0 {
		return false
	}
	if err := auth.ChangeRole(user, "player"); err != nil {
		log.Println(err)
	}
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
		if u.Name == e.Value.(*httpauth.UserData).Name {
			log.Println("duplicate", u.Name)
			return
		}
	}
	q.l.PushBack(u)
}

func (q *UserQueue) PopFront() *httpauth.UserData {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.l.Len() > 0 {
		return q.l.Remove(q.l.Front()).(*httpauth.UserData)
	}
	return nil
}
