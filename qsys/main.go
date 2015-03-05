package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/websocket"

	"h12.me/httpauth"
)

var (
	N = 5

	backend     httpauth.GobFileAuthBackend
	backendfile = "auth.gob"

	auth    httpauth.Authorizer
	waiters UserQueue
	players playerPool

	port = 8009
	tem  *template.Template
)

func main() {
	var err error
	tem, err = template.ParseGlob("template/*.html")
	if err != nil {
		panic(err)
	}
	// create the backend storage, remove when all done
	os.Create(backendfile)
	defer os.Remove(backendfile)
	// create the backend
	backend, err = httpauth.NewGobFileAuthBackend(backendfile)
	if err != nil {
		panic(err)
	}
	auth, err = httpauth.NewAuthorizer(backend, []byte("cookie-encryption-key"), "waiter",
		map[string]httpauth.Role{
			"waiter": 30,
			"player": 50,
			"admin":  100,
		})
	// create a default user
	hash, err := bcrypt.GenerateFromPassword([]byte("adminadmin"), 8)
	if err != nil {
		panic(err)
	}
	defaultUser := httpauth.UserData{Username: "admin", Email: "admin@localhost", Hash: hash, Role: "admin"}
	err = backend.SaveUser(defaultUser)
	if err != nil {
		panic(err)
	}
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
