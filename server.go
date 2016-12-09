package main

import (
	"log"
	"net/http"
)

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	// stub
	log.Println(r)
	w.WriteHeader(http.StatusTeapot)
}

func main() {
	http.HandleFunc("/", handleHTTP)
	http.ListenAndServe("localhost:8080", nil)
}
