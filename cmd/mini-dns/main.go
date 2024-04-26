package main

import (
	"errors"
	"flag"
	"fmt"
	"slices"
	"strings"

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
	multi := make(chan minidns.MultiRequest)
	ca := make(chan minidns.Request)
	go archive.Manage(mongoURI, arch, multi)
	go cache.Run(ca, arch)
	server.Serve(ca, multi)
}

func defaultRun() {
	arch := make(chan minidns.Request)
	multi := make(chan minidns.MultiRequest)
	go archive.Manage(mongoURI, arch, multi)
	server.Serve(arch, multi)

}

func parser(url string) ([]string, error) {
	fmt.Println(url)
	s := strings.Split(url, ".")
	if len(s) < 2 {
		return []string{}, errors.New("invalid url")
	}
	sub := len(s) - 2

	str := strings.Join(s[:sub], ".")
	s = slices.Delete(s, 0, sub)

	slices.Reverse(s)
	s = append(s, str)

	if s[2] == "" {
		s[2] = "-val"
	}
	return s, nil
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
