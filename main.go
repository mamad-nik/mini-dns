package main

import (
	"slices"
	"strings"

	"github.com/mamad-nik/mini-dns/archive"
)

const (
	url     = "www.google.com"
	youtube = "www.youtube.com"
)

/*
{
    sld: 'google',
    subdomains: [
      { domain: 'www', ip: '142.250.179.228' },
      { domain: 'dns', ip: '8.8.8.8' }
    ]
  },
  {
    sld: 'youtube',
    subdomains: [ { domain: 'www', ip: '216.58.213.14' } ]
  },
  {
    sld: 'mamad',
    subdomains: [ { domain: 'www', ip: '54.54.54.54' } ]
  }
*/

func parser(url string) (s []string) {
	s = strings.Split(url, ".")
	sub := len(s) - 2

	str := strings.Join(s[:sub], ".")
	s = slices.Delete(s, 0, sub)

	slices.Reverse(s)
	s = append(s, str)
	return
}

func main() {
	// res, err := agent.LookUp(youtube)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//server.Serve()
	db := archive.NewDB("mongodb://localhost:27017")
	//db.Insert("mamad.com", "54.54.54.55")
	db.Find("mamad.com")

	//fmt.Println(parser("google.com"))
}
