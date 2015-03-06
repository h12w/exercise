package main

import "h12.me/httpauth"

type User struct {
	*httpauth.UserData
	c chan *Message
}

type Message struct {
	Num int
}
