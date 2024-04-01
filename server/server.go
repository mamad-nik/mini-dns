package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Serve() {
	r := mux.NewRouter()

	r.HandleFunc("/dns", input)

	server := http.Server{
		Addr:    ":3535",
		Handler: r,
	}

	server.ListenAndServe()
}
