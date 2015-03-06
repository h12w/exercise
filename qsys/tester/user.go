package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
)

// User simulates the behavior of a user
type User struct {
	Name string
	Pass string
	c    *http.Client // each user must have its own cookie jar
}

// GamePage encapsule the data of a page needed for tests
type GamePage struct {
	Waiting     bool
	UserTotal   int
	UserAheadCh chan int
}

// NewUser creates a user from name and password
func NewUser(name, pass string) *User {
	jar, _ := cookiejar.New(nil)
	return &User{name, pass, &http.Client{Jar: jar}}
}

func (u *User) Login(s *Server) (*GamePage, error) {
	root, err := s.PostForm(u.c, "login", url.Values{
		"username": []string{u.Name},
		"password": []string{u.Pass},
	})
	if err != nil {
		return nil, err
	}
	if root.Html(Id("login")) != nil {
		return nil, errors.New("fail to login")
	} else if sTotal := root.Html(Id("game")).Span(Id("user-total")).Text(); sTotal != nil {
		if total, err := strconv.Atoi(*sTotal); err != nil {
			return nil, err
		} else {
			return &GamePage{UserTotal: total}, nil
		}
	} else if root.Html(Id("wait")) != nil {
		return u.Wait(s)
	}
	// error case
	return nil, errors.New("fail to parse page\n" + *root.Render())
}

func (u *User) Register(s *Server) error {
	_, err := s.PostForm(u.c, "register", url.Values{
		"username": []string{u.Name},
		"password": []string{u.Pass},
	})
	return err
}

func (u *User) Logout(s *Server) error {
	_, err := s.Get(u.c, "logout")
	return err
}

func (u *User) Wait(s *Server) (*GamePage, error) {
	ws, err := s.Dial("wait-num", u.c.Jar)
	if err != nil {
		return nil, err
	}
	ch := make(chan int)
	go func() {
		var buf [512]byte
		for {
			n, err := ws.Read(buf[:])
			if err != nil {
				log.Println(err)
				close(ch)
				break
			}
			userAhead, err := strconv.Atoi(string(buf[:n]))
			if err != nil {
				log.Println(err)
				close(ch)
				break
			}
			ch <- userAhead
		}
	}()
	return &GamePage{
		Waiting:     true,
		UserAheadCh: ch}, nil
}
