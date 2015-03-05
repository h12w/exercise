package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/hailiang/httpauth"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/websocket"
)

func serveWaitNum(ws *websocket.Conn) {
	//io.Copy(ws, ws)
	fmt.Fprintf(ws, "5")
	time.Sleep(time.Second)
	fmt.Fprintf(ws, "2")
	log.Println(ws.Request())
}

var (
	N           = 5
	backend     httpauth.GobFileAuthBackend
	au          httpauth.Authorizer
	roles       map[string]httpauth.Role
	port        = 8009
	backendfile = "auth.gob"
	tem         *template.Template
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
	// create some default roles
	roles = make(map[string]httpauth.Role)
	roles["waiter"] = 30
	roles["player"] = 50
	roles["admin"] = 80
	au, err = httpauth.NewAuthorizer(backend, []byte("cookie-encryption-key"), "player", roles)
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
	r.HandleFunc("/admin", handleAdmin).Methods("GET")
	r.HandleFunc("/add_user", postAddUser).Methods("POST")
	r.HandleFunc("/change", postChange).Methods("POST")
	r.HandleFunc("/", handleGame).Methods("GET") // authorized page
	r.HandleFunc("/logout", handleLogout)
	http.Handle("/wait-num", websocket.Handler(serveWaitNum))
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))
	http.Handle("/", r)
	fmt.Printf("Server running on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
func getLogin(rw http.ResponseWriter, req *http.Request) {
	messages := au.Messages(rw, req)
	fmt.Fprintf(rw, `
<html>
<head><title>Login</title></head>
<body>
<h1>Httpauth example</h1>
<h2>Entry Page</h2>
<p><b>Messages: %v</b></p>
<h3>Login</h3>
<form action="/login" method="post" id="login">
<input type="text" name="username" placeholder="username"><br>
<input type="password" name="password" placeholder="password"></br>
<button type="submit">Login</button>
</form>
<h3>Register</h3>
<form action="/register" method="post" id="register">
<input type="text" name="username" placeholder="username"><br>
<input type="password" name="password" placeholder="password"></br>
<input type="email" name="email" placeholder="email@example.com"></br>
<button type="submit">Register</button>
</form>
</body>
</html>
`, messages)
}
func postLogin(rw http.ResponseWriter, req *http.Request) {
	username := req.PostFormValue("username")
	password := req.PostFormValue("password")
	if err := au.Login(rw, req, username, password, "/"); err != nil && err.Error() == "already authenticated" {
		http.Redirect(rw, req, "/", http.StatusSeeOther)
	} else if err != nil {
		fmt.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	}
}
func postRegister(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Username = req.PostFormValue("username")
	user.Email = req.PostFormValue("email")
	password := req.PostFormValue("password")
	if err := au.Register(rw, req, user, password); err == nil {
		postLogin(rw, req)
	} else {
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
	}
}
func postAddUser(rw http.ResponseWriter, req *http.Request) {
	var user httpauth.UserData
	user.Username = req.PostFormValue("username")
	user.Email = req.PostFormValue("email")
	password := req.PostFormValue("password")
	user.Role = req.PostFormValue("role")
	if err := au.Register(rw, req, user, password); err != nil {
		// maybe something
	}
	http.Redirect(rw, req, "/admin", http.StatusSeeOther)
}
func postChange(rw http.ResponseWriter, req *http.Request) {
	email := req.PostFormValue("new_email")
	au.Update(rw, req, "", email)
	http.Redirect(rw, req, "/", http.StatusSeeOther)
}
func handleGame(rw http.ResponseWriter, req *http.Request) {
	if err := au.AuthorizeRole(rw, req, "player", true); err != nil {
		fmt.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
		return
	}
	if user, err := au.CurrentUser(rw, req); err == nil {
		type data struct {
			User    httpauth.UserData
			WaitNum int
		}
		d := data{User: user}
		if err := tem.ExecuteTemplate(rw, "game.html", d); err != nil {
			log.Println(err.Error())
		}
	}
}
func handleAdmin(rw http.ResponseWriter, req *http.Request) {
	if err := au.AuthorizeRole(rw, req, "admin", true); err != nil {
		fmt.Println(err)
		http.Redirect(rw, req, "/login", http.StatusSeeOther)
		return
	}
	if user, err := au.CurrentUser(rw, req); err == nil {
		type data struct {
			User  httpauth.UserData
			Roles map[string]httpauth.Role
			Users []httpauth.UserData
			Msg   []string
		}
		messages := au.Messages(rw, req)
		users, err := backend.Users()
		if err != nil {
			panic(err)
		}
		d := data{User: user, Roles: roles, Users: users, Msg: messages}
		t, err := template.New("admin").Parse(`
<html>
<head><title>Admin page</title></head>
<body>
<h1>Httpauth example<h1>
<h2>Admin Page</h2>
<p>{{.Msg}}</p>
{{ with .User }}<p>Hello {{ .Username }}, your role is '{{ .Role }}'. Your email is {{ .Email }}.</p>{{ end }}
<p><a href="/">Back</a> <a href="/logout">Logout</a></p>
<h3>Users</h3>
<ul>{{ range .Users }}<li>{{.Username}}</li>{{ end }}</ul>
<form action="/add_user" method="post" id="add_user">
<h3>Add user</h3>
<p><input type="text" name="username" placeholder="username"><br>
<input type="password" name="password" placeholder="password"><br>
<input type="email" name="email" placeholder="email"><br>
<select name="role">
<option value="">role<option>
{{ range $key, $val := .Roles }}<option value="{{$key}}">{{$key}} - {{$val}}</option>{{ end }}
</select></p>
<button type="submit">Submit</button>
</form>
</body>
`)
		if err != nil {
			panic(err)
		}
		t.Execute(rw, d)
	}
}
func handleLogout(rw http.ResponseWriter, req *http.Request) {
	if err := au.Logout(rw, req); err != nil {
		fmt.Println(err)
		// this shouldn't happen
		return
	}
	http.Redirect(rw, req, "/", http.StatusSeeOther)
}
