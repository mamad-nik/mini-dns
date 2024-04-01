package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/mamad-nik/mini-dns/agent"
)

const (
	url = "www.google.com"
)

func parser(url string) (s []string) {
	s = strings.Split(url, ".")
	sub := len(s) - 2

	str := strings.Join(s[:sub], ".")
	s = slices.Delete(s, 0, sub)

	slices.Reverse(s)
	s = append(s, str)
	//s = slices.Delete(s, 0, 1)
	return
}

func main() {
	fmt.Println(parser(url))
	if res, err := agent.LookUp(url); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(res)
	}
	//server.Serve()
	// db := archive.NewDB("mongodb://localhost:27017")
	// db.Find("www.google.com")
}
