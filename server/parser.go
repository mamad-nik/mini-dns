package server

import (
	"encoding/json"
	"fmt"
	"net/http"
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

func input(w http.ResponseWriter, r *http.Request) {
	var ij inputJson
	err := json.NewDecoder(r.Body).Decode(&ij)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsoniseErr("not a valid json request", w)
	}

	fmt.Printf("r.URL.Path: %v\n", r.RemoteAddr)
}
