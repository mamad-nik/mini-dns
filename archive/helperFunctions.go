package archive

import (
	"slices"
	"strings"
)

func parser(url string) (s []string) {
	s = strings.Split(url, ".")
	sub := len(s) - 2

	str := strings.Join(s[:sub], ".")
	s = slices.Delete(s, 0, sub)

	slices.Reverse(s)
	s = append(s, str)

	if s[2] == "" {
		s[2] = "-val"
	}
	return
}
