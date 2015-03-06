package main

import "flag"

type Option struct {
	ServerURL string
}

var opt Option

func init() {
	flag.StringVar(&opt.ServerURL, "url", "http://127.0.0.1:9009", "Server URL")
}

func main() {
	flag.Parse()
	s, _ := NewServer(opt.ServerURL)
	u := NewUser("c", "c")
	u.Register(s)

	//u := NewUser("b", "b")
	//page, err := u.Login(s)
	//if err != nil {
	//	log.Println(err)
	//	os.Exit(1)
	//}
	//fmt.Println(page)
	//if page.Waiting {
	//	fmt.Println("ahead", <-page.UserAheadCh)
	//}
	//err = u.Logout(s)
	//if err != nil {
	//	log.Println(err)
	//	os.Exit(1)
	//}
}
