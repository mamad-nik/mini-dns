package archive

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

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

func reconstruct(tld, sld, sub string) string {
	if sub == "-val" {
		sub = ""
	} else {
		sub += "."
	}
	return sub + sld + "." + tld
}
