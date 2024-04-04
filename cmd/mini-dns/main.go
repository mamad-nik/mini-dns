package main

import (
	"github.com/mamad-nik/mini-dns/archive"
)

const (
	google   = "www.google.com"
	youtube  = "www.youtube.com"
	mongoURI = "mongodb://localhost:27017"
)

func main() {
	/*
		ch := make(chan minidns.Request)
		go archive.Manage(, ch)
		res := make(chan string)
		err := make(chan error)

		ch <- minidns.Request{
			Domain: "www.jadi.net",
			IP:     res,
			Err:    err,
		}

		select {
		case e := <-err:
			fmt.Println(e)
			return
		case ip := <-res:
			fmt.Println(ip)
			return
		}
	*/
	db := archive.NewDB(mongoURI)
	db.Restore()
}
