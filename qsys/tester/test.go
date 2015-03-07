package main

import (
	"fmt"
	"log"
	"os"
)

// autoTest logic:
// 1. restart server
// 2. it can only be tested with a single tester instance
func test(auto bool) {
	s, _ := NewServer(opt.ServerURL)
	users := getUsers() // users must be reused so that cookies are preserved
	pass := true
	var waitCount int
	for j := 0; j < 2; j++ { // log out and log in again to prove users are indeed logged out
		for i, user := range users {
			page, err := user.Login(s)
			if err != nil {
				log.Println(err)
				os.Exit(1)
			}
			log.Printf("user: %s, waiting: %v", user.Name, page.Waiting)
			if page.Waiting {
				waitCount = <-page.UserAheadCh
				log.Printf("%d users waiting ahead", waitCount)
			} else {
				log.Printf("player count %d", page.PlayerCount)
			}
			if auto {
				if i < opt.ServerCap {
					if page.Waiting {
						log.Println("ERROR: wait before reaching capacity")
						pass = false
					}
					if page.PlayerCount != (i + 1) {
						log.Println("ERROR: player count not correct")
						pass = false
					}
				} else {
					if !page.Waiting {
						log.Println("ERROR: not wait after reaching capacity")
						pass = false
					}
					if waitCount != (i - opt.ServerCap + 1) {
						log.Println("ERROR: count of users ahead is not correct")
						pass = false
					}
				}
			}
		}

		for _, user := range users {
			user.Logout(s)
		}
	}

	if auto && pass {
		log.Println("TEST PASS.")
	}
}

func getUsers() []*User {
	users := make([]*User, opt.UserCount)
	for i := 0; i < opt.UserCount; i++ {
		name := fmt.Sprintf("u%d", i)
		users[i] = NewUser(name, name)
	}
	return users
}
