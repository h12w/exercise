package main

import (
	"net/http"
	"net/url"

	"golang.org/x/net/websocket"
	"h12.me/html-query"
)

type Server struct {
	URL *url.URL
}

func NewServer(host string) (*Server, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}
	return &Server{u}, nil
}

func (s *Server) Get(c *http.Client, relativeURL string) (*query.Node, error) {
	resp, err := c.Get(s.URL.String() + relativeURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return query.Parse(resp.Body)
}

func (s *Server) PostForm(c *http.Client, relativeURL string, data url.Values) (*query.Node, error) {
	resp, err := c.PostForm(s.URL.String()+relativeURL, data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return query.Parse(resp.Body)
}

func (s *Server) Dial(relativeURL string, jar http.CookieJar) (*websocket.Conn, error) {
	u := *s.URL
	u.Scheme = "ws"
	u.Path += relativeURL
	config, err := websocket.NewConfig(u.String(), s.URL.String())
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("", u.String(), nil)
	if err != nil {
		return nil, err
	}
	for _, c := range jar.Cookies(s.URL) {
		req.AddCookie(c)
	}
	config.Header = req.Header
	return websocket.DialConfig(config)
}
