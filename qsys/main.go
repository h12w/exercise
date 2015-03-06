package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/websocket"

	"h12.me/httpauth"
)

type Option struct {
	Port         string
	MaxCapacity  int
	CookieKey    string
	TemplateDir  string
	DbType       string
	DbSource     string
	GenUserCount int
}

var opt Option

func init() {
	flag.StringVar(&opt.Port, "port", "9009", "server port")
	flag.IntVar(&opt.MaxCapacity, "cap", 100, "maximum player count")
	flag.StringVar(&opt.CookieKey, "ckey", "cookie-encryption-key", "cookie key")
	flag.StringVar(&opt.TemplateDir, "tdir", "template", "template directory")
	flag.StringVar(&opt.DbType, "dbtype", "mem", "database type: mem, mongo, sql driver")
	flag.StringVar(&opt.DbSource, "dbsrc", "", "database source")
	flag.IntVar(&opt.GenUserCount, "gen", 0, "how many fake users to generate (debug only)")
}

var (
	backend httpauth.AuthBackend
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
	switch opt.DbType {
	case "mem":
		backend, err = httpauth.NewMemAuthBackend()
	case "mongo":
		backend, err = httpauth.NewMongoAuthBackend(opt.DbSource, "auth")
	default:
		backend, err = httpauth.NewSQLAuthBackend(opt.DbType, opt.DbSource)
	}
	if err != nil {
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
	log.Println("generating", opt.GenUserCount, "users")
	for i := 0; i < opt.GenUserCount; i++ {
		name := fmt.Sprintf("u%d", i)
		pass := name
		hash, _ := bcrypt.GenerateFromPassword([]byte(name+pass), 4)
		_ = hash
		user := httpauth.UserData{
			Name: name,
			Hash: hash,
			Role: "player",
		}
		backend.SaveUser(user)
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
