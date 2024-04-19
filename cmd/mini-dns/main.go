package main

import (
	"flag"

	minidns "github.com/mamad-nik/mini-dns"
	"github.com/mamad-nik/mini-dns/archive"
	"github.com/mamad-nik/mini-dns/cache"
	"github.com/mamad-nik/mini-dns/server"
)

const (
	google   = "www.google.com"
	youtube  = "www.youtube.com"
	mongoURI = "mongodb://localhost:27017"
)

func cachedRun() {
	arch := make(chan minidns.Request)
	ca := make(chan minidns.Request)
	go archive.Manage(mongoURI, arch)
	go cache.Run(ca, arch)
	server.Serve(ca)
}

func defaultRun() {
	arch := make(chan minidns.Request)
	go archive.Manage(mongoURI, arch)
	server.Serve(arch)
}

func main() {
	cf := flag.Bool("cache", false, "should i use cache?")
	flag.Parse()

	if *cf {
		cachedRun()
	} else {
		defaultRun()
	}
}
