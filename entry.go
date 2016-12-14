package main

import "time"

type EntryType int

const (
	None EntryType = iota
	Comment
	Question
)

func (e EntryType) String() string {
	switch e {
	case Comment:
		return "c"
	case Question:
		return "q"
	}
	return ""
}

func (e EntryType) MarshalText() ([]byte, error) {
	switch e {
	case Comment:
		return []byte("c"), nil
	case Question:
		return []byte("q"), nil
	}
	return []byte{}, nil
}

func (e *EntryType) UnmarshalText(b []byte) error {
	switch string(b) {
	case "c":
		*e = Comment
	case "q":
		*e = Question
	default:
		*e = None
	}
	return nil
}

type Entry struct {
	ID        int       `json:"id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	Type      EntryType `json:"type"` // Question or Comment
	Score     int       `json:"score"`
	Voted     bool      `json:"voted"` // Only used for output
	Timestamp time.Time `json:"timestamp"`

	votes map[string]int
}

func (e Entry) HasVoted(u string) bool {
	if _, ok := e.votes[u]; ok {
		return true
	}
	return false
}

func (e *Entry) Vote(u string, v int) (score int) {
	if _, ok := e.votes[u]; !ok {
		e.votes[u] = v
		e.Score += v
	}
	return e.Score
}

func LessByScore(e1, e2 Entry) bool {
	if e1.Score < e2.Score {
		return true
	} else if e1.Score == e2.Score {
		return e1.Timestamp.Before(e2.Timestamp)
	}
	return false
}

func GreaterByScore(e1, e2 Entry) bool {
	if e1.Score > e2.Score {
		return true
	} else if e1.Score == e2.Score {
		return e1.Timestamp.Before(e2.Timestamp)
	}
	return false
}

func NewEntry(id int, author, text string, etype EntryType) Entry {
	e := Entry{}
	e.ID = id
	e.Author = author
	e.Text = text
	e.Type = etype

	e.Score = 0
	e.Timestamp = time.Now()
	e.votes = make(map[string]int)
	return e
}

type Entries struct {
	E  []Entry
	By func(e1, e2 Entry) bool
}

func (e Entries) Len() int {
	return len(e.E)
}

func (e Entries) Swap(i, j int) {
	e.E[i], e.E[j] = e.E[j], e.E[i]
}

func (e Entries) Less(i, j int) bool {
	return e.By(e.E[i], e.E[j])
}
