package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

type Server struct {
	URL string
}

func NewServer(url string) *Server {
	return &Server{url}
}

func (s *Server) PostForm(c *http.Client, relativeURL string, data url.Values) error {
	resp, err := c.PostForm(s.URL+relativeURL, data)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_ = resp
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Println(resp.Header)
	log.Println(string(buf))
	return err
}
