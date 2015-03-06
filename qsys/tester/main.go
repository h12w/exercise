package main

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
)

type User struct {
	Name string
	Pass string
	c    *http.Client // each user must have its own cookie jar
}

func NewUser(name, pass string) *User {
	jar, _ := cookiejar.New(nil)
	return &User{name, pass, &http.Client{Jar: jar}}
}

func (u *User) Login(s *Server) error {
	return s.PostForm(u.c, "login", url.Values{
		"username": []string{u.Name},
		"password": []string{u.Pass},
	})
}

func main() {
	s := NewServer("http://127.0.0.1:8009/")
	u := NewUser("a", "a")
	if err := u.Login(s); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
