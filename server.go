package main

import (
	"log"
	"net/http"
)

type Server struct {
	Streams map[string]Stream
}

//Streams map[string]Stream

func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	// stub
	switch r.Method {
	case http.MethodGet:
		params := r.URL.Query()
		log.Println("GET", params)

	case http.MethodPost:
		r.ParseForm()
		params := r.Form
		if comment, ok := params["comment"]; ok {
			log.Println("POST", comment)
		} else if question, ok := params["question"]; ok {
			log.Println("POST", question)
		} else if question_id, ok := params["question_id"]; ok {
			log.Println("POST", question_id)
		}
	}
	w.WriteHeader(http.StatusTeapot)
}

func main() {
	s := Server{map[string]Stream{}}
	http.HandleFunc("/video/", s.handleHTTP)
	http.ListenAndServe("localhost:8080", nil)
}
