package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	similarAddr string
	sentiAddr   string
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
	case "":
		return Question
	}
	return Question
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
	text := params.Get("text")
	hc := http.Client{}

	form := url.Values{}
	form.Add("text", text)

	req, _ := http.NewRequest("GET", similarAddr+"/similar", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := hc.Do(req)
	log.Println(similarAddr, text, resp, err)
	b := []byte{}
	resp.Body.Read(b)
	w.WriteHeader(resp.StatusCode)
	w.Write(b)
}

func entryToForm(e Entry) url.Values {
	form := url.Values{}
	form.Add("id", strconv.Itoa(e.ID))
	form.Add("author", e.Author)
	form.Add("text", e.Text)
	form.Add("type", string(e.Type))
	form.Add("score", strconv.Itoa(e.Score))
	form.Add("timestamp", strconv.FormatInt(e.Timestamp.Unix(), 10))
	if e.Voted {
		form.Add("voted", "true")
	} else {
		form.Add("voted", "false")
	}

	return form
}

func sentToOthers(e Entry, addr string) {
	hc := http.Client{}

	form := entryToForm(e)

	req, _ := http.NewRequest("POST", addr+"/add", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := hc.Do(req)
	log.Println(addr, e, resp, err)
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
	m := Message{ReplyChan: ch, Type: Add, E: Entry{Author: author, Text: text, Type: t}}
	s.s.InputChannel() <- m
	m = <-ch

	log.Println(m)
	if m.Type != Error {
		w.Write([]byte(fmt.Sprintf("%d", m.E.ID)))
		go sentToOthers(m.E, similarAddr)
		sentToOthers(m.E, sentiAddr)
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
		vote = 1
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

func main() {
	var (
		host = flag.String("host", "", "host address to listen on")
		port = flag.String("port", "8080", "port to listen on")
	)
	flag.StringVar(&sentiAddr, "senti-addr", "", "address of sentimental analysis server")
	flag.StringVar(&similarAddr, "similar-addr", "", "address of similarity analysis server")
	flag.Parse()
	sentiAddr = "http://" + sentiAddr
	similarAddr = "http://" + similarAddr

	s := MakeServer()
	http.HandleFunc("/all", s.GetAll)
	http.HandleFunc("/recent", s.GetRecent)
	http.HandleFunc("/top", s.GetTop)
	http.HandleFunc("/similar", s.GetSimilar)
	http.HandleFunc("/add", s.Add)
	http.HandleFunc("/vote", s.Vote)
	http.HandleFunc("/sentiment", s.Vote)
	http.ListenAndServe(*host+":"+*port, nil)
}
