package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"

	"h12.me/httpauth"
)

type Option struct {
	Port        string
	MaxCapacity int
	CookieKey   string
	TemplateDir string
	DbConn      string
}

var opt Option

func init() {
	flag.StringVar(&opt.Port, "port", "9009", "server port")
	flag.IntVar(&opt.MaxCapacity, "cap", 1, "maximum player count")
	flag.StringVar(&opt.CookieKey, "ckey", "cookie-encryption-key", "cookie key")
	flag.StringVar(&opt.TemplateDir, "tdir", "template", "template directory")
	flag.StringVar(&opt.DbConn, "dbc", "db/auth.gob", "database connection string")
}

var (
	backend httpauth.GobFileAuthBackend
	auth    httpauth.Authorizer
	waiters UserQueue
	players = NewPlayerPool()
	tem     *template.Template
)

func main() {
	flag.Parse()
	var err error
	players.capacity = opt.MaxCapacity
	if tem, err = template.ParseGlob(path.Join(opt.TemplateDir, "*.html")); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if backend, err = httpauth.NewGobFileAuthBackend(opt.DbConn); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if auth, err = httpauth.NewAuthorizer(
		backend,
		[]byte(opt.CookieKey), "player",
		map[string]httpauth.Role{"player": 20}); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	r := mux.NewRouter()
	r.HandleFunc("/login", getLogin).Methods("GET")
	r.HandleFunc("/login", postLogin).Methods("POST")
	r.HandleFunc("/register", postRegister).Methods("POST")
	r.HandleFunc("/", handleGame).Methods("GET") // authorized page
	r.HandleFunc("/logout", handleLogout)
	http.Handle("/wait-num", websocket.Handler(serveWaitNum))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	http.Handle("/", r)
	log.Printf("Server listening on :%s\n", opt.Port)
	http.ListenAndServe(":"+opt.Port, nil)
}
