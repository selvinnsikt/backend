package model

import "github.com/gorilla/websocket"

// NewPlayer is used by both /newGameRoom and /joinGameRoom
type NewPlayer struct {
	Name  string `json:"name"`
	HubID string `json:"hubID"`
}

type PlayerConnection struct {
	Name string
	Conn *websocket.Conn
}

type Message struct {
	Player string `json:"player,omitempty"` // name of player who sent the message
	Text   string `json:"text"`
}
// Example json-object: {"payloadtype":"AnswerFromPlayer","question":1,"votes":["aksel","alf"]}
type Payload struct {
	PayloadType string `json:"payloadtype"`
}

type AnswerFromPlayer struct {
	// Which question number 1-4
	Question int `json:"question"`
	// Votes of players choosen on that question
	// Example: {"Aksel","Alf"]    OR    {"Aksel","Aksel"}
	Votes []string `json:"votes"`
}
