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

func (p *playerPool) Count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return maxCapacity - p.capacity
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
type userNotifier struct {
	*httpauth.UserData
	c chan int
}

func (q *UserQueue) PushBack(u *httpauth.UserData) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for e := q.l.Front(); e != nil; e = e.Next() {
		if u.Name == e.Value.(*userNotifier).Name {
			log.Println("duplicate", u.Name)
			return
		}
	}
	q.l.PushBack(&userNotifier{UserData: u})
}

func (q *UserQueue) PopFront() *httpauth.UserData {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.l.Len() > 0 {
		noti := q.l.Remove(q.l.Front()).(*userNotifier)
		noti.c <- 0
		close(noti.c)
		i := 0
		log.Println("NotifyAll")
		for e := q.l.Front(); e != nil; e = e.Next() {
			noti := e.Value.(*userNotifier)
			log.Println(noti.Name, i)
			noti.c <- i
			i++
		}
		return noti.UserData
	}
	return nil
}

func (q *UserQueue) Register(userName string, c chan int) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	i := 1
	for e := q.l.Front(); e != nil; e = e.Next() {
		if userName == e.Value.(*userNotifier).Name {
			e.Value.(*userNotifier).c = c
			return i
		}
		i++
	}
	return 0
}
