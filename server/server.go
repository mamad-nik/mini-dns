package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	minidns "github.com/mamad-nik/mini-dns"
)

type inputJson struct {
	Url string `json:"url"`
}

type errorJson struct {
	Response string `json:"response"`
}

func jsoniseErr(returnedErr string, w http.ResponseWriter) {
	ej := errorJson{
		Response: returnedErr,
	}
	if err := json.NewEncoder(w).Encode(ej); err != nil {
		return
	}
}

func Serve(reqchan chan minidns.Request) {
	r := mux.NewRouter()

	r.HandleFunc("/dns", func(w http.ResponseWriter, r *http.Request) {
		var ij inputJson
		err := json.NewDecoder(r.Body).Decode(&ij)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			jsoniseErr("not a valid json request", w)
		}

		req := minidns.Request{
			Domain: ij.Url,
			IP:     make(chan string),
			Err:    make(chan error),
		}
		reqchan <- req
		select {
		case e := <-req.Err:
			w.WriteHeader(http.StatusInternalServerError)
			jsoniseErr(e.Error(), w)
		case ip := <-req.IP:
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("%s\n", ip)))

		}
	})

	server := http.Server{
		Addr:    ":3535",
		Handler: r,
	}

	server.ListenAndServe()
}
