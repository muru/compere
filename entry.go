package main

type EntryType int

const (
	Comment EntryType = iota
	Question
	Duplicate
)

type Entry struct {
	ID        int
	StreamID  string // ID for live stream
	Author    string
	Text      string
	EntryType EntryType // Question or Comment
	Score     int
	Deleted   bool

	votes map[string]int
}

func (e *Entry) Vote(u string, v int) (score int) {
	if _, ok := e.votes[u]; !ok {
		e.votes[u] = v
		e.Score += v
	}
	return e.Score
}

func NewEntry(id int, streamid, author, text string, etype EntryType) Entry {
	e := Entry{}
	e.ID = id
	e.StreamID = streamid
	e.Author = author
	e.Text = text
	e.EntryType = etype
	e.Score = 0
	e.votes = make(map[string]int)
	return e
}
