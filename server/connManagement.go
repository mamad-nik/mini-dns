package server

import "net/http"

type conn struct {
	w http.ResponseWriter
	r *http.Request
}

type openConns struct {
	IDs    map[int]*conn
	lastID int
}

func addConn(w http.ResponseWriter, r *http.Request) int {
	/*c := conn{
		w: w,
		r: r,
	}*/
	return 0

}
