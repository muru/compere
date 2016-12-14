package main

import (
	"log"
	"sort"
	"time"
)

const (
	Guest string = "GUEST"
)

type MessageType int

const (
	Error MessageType = iota
	Add
	Vote
)

func (m MessageType) String() string {
	switch m {
	case Add:
		return "Add"
	case Vote:
		return "Vote"
	}
	return "Error"
}

type Message struct {
	Type      MessageType
	E         Entry
	ReplyChan chan<- Message
}

type Stream interface {
	InputChannel() chan<- Message
	GetEntriesByTime(e EntryType, d time.Time) []Entry
	GetEntriesByScore(e EntryType, num int) []Entry
	Close()
}

type MapStream struct {
	entries []Entry
	input   chan Message
	exit    chan bool
}

func (s *MapStream) handleChanges(m Message) {
	defer func() {
		if x := recover(); x != nil {
			m.Type = Error
			log.Println(m, x)
		}
		m.ReplyChan <- m
	}()
	switch m.Type {
	case Add:
		id := len(s.entries)
		e := NewEntry(id, m.E.Author, m.E.Text, m.E.Type)
		s.entries = append(s.entries, e)
		m.E = e
	case Vote:
		m.E.Score = s.entries[m.E.ID].Vote(m.E.Author, m.E.Score)
	}
}

func (s *MapStream) acceptInput() {
	for {
		select {
		case m := <-s.input:
			s.handleChanges(m)
		case <-s.exit:
			return
		}
	}
}

func (s *MapStream) InputChannel() chan<- Message {
	return s.input
}

func compareType(t1, t2 EntryType) bool {
	return t1 == None || t2 == None || t1 == t2
}

func (s MapStream) GetEntriesByTime(e EntryType, t time.Time) (es []Entry) {
	defer func() {
		if x := recover(); x != nil {
			es = make([]Entry, 0)
		}
	}()
	es = make([]Entry, 0)
	for _, v := range s.entries {
		if t.Before(v.Timestamp) && compareType(v.Type, e) {
			es = append(es, v)
		}
	}
	return es
}

func (s *MapStream) GetEntriesByScore(e EntryType, num int) (es []Entry) {
	defer func() {
		if x := recover(); x != nil {
			es = nil
		}
	}()
	out := &Entries{}
	out.By = GreaterByScore
	for _, v := range s.entries {
		if compareType(e, v.Type) {
			out.E = append(out.E, v)
		}
	}
	if num == -1 || num > len(out.E) {
		num = len(out.E)
	}
	sort.Sort(out)
	return out.E[:num]
}

func (s *MapStream) Close() {
	s.exit <- true
}

func MakeStream() Stream {
	s := MapStream{}
	s.exit = make(chan bool)
	s.input = make(chan Message)
	s.entries = []Entry{}
	go s.acceptInput()
	return &s
}
