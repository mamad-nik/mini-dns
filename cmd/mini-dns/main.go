package main

import (
	"fmt"

	minidns "github.com/mamad-nik/mini-dns"
	"github.com/mamad-nik/mini-dns/archive"
)

const (
	google  = "www.google.com"
	youtube = "www.youtube.com"
)

func main() {
	ch := make(chan minidns.Dn)
	go archive.Manager("mongodb://localhost:27017", ch)
	res := make(chan string)
	err := make(chan error)

	ch <- minidns.Dn{
		Domain: "google.com",
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
}
