package main

import "flag"

type Option struct {
	ServerURL string
	ServerCap int
	UserCount int
	Auto      bool
}

var opt Option

func init() {
	flag.StringVar(&opt.ServerURL, "url", "http://127.0.0.1:9009/", "Server URL")
	flag.IntVar(&opt.ServerCap, "cap", 100, "Server capacity")
	flag.IntVar(&opt.UserCount, "cnt", 200, "Total users to login")
	flag.BoolVar(&opt.Auto, "auto", true, "automatic checking or not")
}

func main() {
	flag.Parse()
	test(opt.Auto)
}
