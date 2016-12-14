package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Server struct {
	s Stream
}

func MakeServer() *Server {
	s := Server{}
	s.s = MakeStream()
	return &s
}

func getType(u url.Values) EntryType {
	switch u.Get("type") {
	case "q":
		return Question
	case "c":
		return Comment
	}
	return None
}

func (s *Server) GetAll(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(x)
		}
	}()
	params := r.URL.Query()
	log.Println(params)
	author := params.Get("author")
	t := getType(params)
	ret := s.s.GetEntriesByTime(t, time.Unix(0, 0))
	for k, _ := range ret {
		ret[k].Voted = ret[k].HasVoted(author)
	}
	out, ok := json.Marshal(ret)
	if ok == nil {
		w.Write(out)
	} else {
		log.Panicln("Could not JSON: ", ok)
	}
}

func (s *Server) GetRecent(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	params := r.URL.Query()
	log.Println(params)
	author := params.Get("author")
	t := getType(params)
	ret := s.s.GetEntriesByTime(t, time.Now().Add(-10*time.Minute))
	for k, _ := range ret {
		ret[k].Voted = ret[k].HasVoted(author)
	}
	out, ok := json.Marshal(ret)
	if ok == nil {
		w.Write(out)
	} else {
		log.Panicln("Could not JSON: ", ok)
	}
}

func (s *Server) GetTop(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	params := r.URL.Query()
	log.Println(params)
	author := params.Get("author")
	t := getType(params)
	ret := s.s.GetEntriesByScore(t, 100)
	for k, _ := range ret {
		ret[k].Voted = ret[k].HasVoted(author)
	}
	out, ok := json.Marshal(ret)
	if ok == nil {
		w.Write(out)
	} else {
		log.Panicln("Could not JSON: ", ok)
	}
}

func (s *Server) GetSimilar(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	params := r.URL.Query()
	log.Println(params)
	//t := getType(params)
	//text := params.Get("text")
	w.WriteHeader(http.StatusTeapot)
}

func (s *Server) Add(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	r.ParseForm()
	params := r.Form
	author := params.Get("author")
	t := getType(params)
	text := params.Get("text")
	if t == None {
		t = Comment
	}
	ch := make(chan Message)
	m := Message{ReplyChan: ch, Type: Add, E: Entry{Author: author, Text: text, EntryType: t}}
	s.s.InputChannel() <- m
	m = <-ch

	log.Println(m)
	if m.Type != Error {
		w.Write([]byte(fmt.Sprintf("%d", m.E.ID)))
	} else {
		log.Panicln(m)
	}
}

func (s *Server) Vote(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if x := recover(); x != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	r.ParseForm()
	params := r.Form
	log.Println(params)

	author := params.Get("author")
	id, err := strconv.Atoi(params.Get("id"))
	if err != nil {
		log.Panicln("Could not parse: ", params.Get("id"))
	}
	vote, err := strconv.Atoi(params.Get("vote"))
	if err != nil {
		log.Panicln("Could not parse: ", params.Get("vote"))
	}

	ch := make(chan Message)
	m := Message{ReplyChan: ch, Type: Vote, E: Entry{Author: author, ID: id, Score: vote}}
	s.s.InputChannel() <- m

	m = <-ch
	if m.Type != Error {
		w.Write([]byte(fmt.Sprintf("%d", m.E.Score)))
	} else {
		log.Panicln(m)
	}
}

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
	s := MakeServer()
	http.HandleFunc("/all", s.GetAll)
	http.HandleFunc("/recent", s.GetRecent)
	http.HandleFunc("/top", s.GetTop)
	http.HandleFunc("/similar", s.GetSimilar)
	http.HandleFunc("/add", s.Add)
	http.HandleFunc("/vote", s.Vote)
	http.HandleFunc("/sentiment", s.Vote)
	http.ListenAndServe("localhost:8080", nil)
}
