package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/net/websocket"

	"h12.me/httpauth"
)

var (
	maxCapacity = 2

	backend     httpauth.GobFileAuthBackend
	backendfile = "db/auth.gob"

	auth    httpauth.Authorizer
	waiters UserQueue
	players = NewPlayerPool()

	cookieKey = []byte("cookie-encryption-key")
	port      = 8009
	tem       *template.Template
)

func main() {
	players.capacity = maxCapacity
	var err error
	tem, err = template.ParseGlob("template/*.html")
	if err != nil {
		panic(err)
	}
	// create the backend storage, remove when all done
	//os.Create(backendfile)
	// create the backend
	backend, err = httpauth.NewGobFileAuthBackend(backendfile)
	if err != nil {
		panic(err)
	}
	auth, err = httpauth.NewAuthorizer(backend, cookieKey, "waiter",
		map[string]httpauth.Role{
			"waiter": 10,
			"player": 20,
		})
	// set up routers and route handlers
	r := mux.NewRouter()
	r.HandleFunc("/login", getLogin).Methods("GET")
	r.HandleFunc("/login", postLogin).Methods("POST")
	r.HandleFunc("/register", postRegister).Methods("POST")
	//r.HandleFunc("/admin", handleAdmin).Methods("GET")
	r.HandleFunc("/", handleGame).Methods("GET") // authorized page
	r.HandleFunc("/logout", handleLogout)
	http.Handle("/wait-num", websocket.Handler(serveWaitNum))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	http.Handle("/", r)
	fmt.Printf("Server running on port %d\n", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
