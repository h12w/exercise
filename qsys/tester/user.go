package main

import (
	"errors"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"

	"golang.org/x/net/websocket"

	"h12.io/html-query/expr"
)

var (
	Id = expr.Id
)

// User simulates the behavior of a user
type User struct {
	Name string
	Pass string
	c    *http.Client // each user must have its own cookie jar
	ws   *websocket.Conn
}

// GamePage encapsule the data of a page needed for tests
type GamePage struct {
	Waiting     bool
	PlayerCount int
	UserAheadCh chan int
}

// NewUser creates a user from name and password
func NewUser(name, pass string) *User {
	jar, _ := cookiejar.New(nil)
	return &User{Name: name, Pass: pass, c: &http.Client{Jar: jar}}
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
			return &GamePage{PlayerCount: total}, nil
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
	if u.ws != nil {
		u.ws.Close()
		u.ws = nil
	}
	return err
}

func (u *User) Wait(s *Server) (*GamePage, error) {
	ws, err := s.Dial("wait-num", u.c.Jar)
	u.ws = ws
	if err != nil {
		return nil, err
	}
	ch := make(chan int)
	go func() {
		defer ws.Close()
		var buf []byte
		for {
			err := websocket.Message.Receive(ws, &buf)
			if err != nil {
				log.Println(err)
				close(ch)
				break
			}
			userAhead, err := strconv.Atoi(string(buf))
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
