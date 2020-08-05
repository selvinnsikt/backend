package model

import "golang.org/x/net/websocket"

// NewPlayer is used by both /newGameRoom and /joinGameRoom
type NewPlayer struct {
	Name       string `json:"name"`
	GameRoomID string `json:"game_room_id"`
}

type NewConnection struct {
	Name          string
	addClientChan chan *websocket.Conn
}

type Message struct {
	Text string `json:"text"`
}

type Answer struct {
	// Which question number 1-4
	Question int `json:"question"`
	// Votes of players choosen on that question
	// Example: {"Aksel","Alf"]    OR    {"Aksel","Aksel"}
	Votes []string `json:"votes"`
}
