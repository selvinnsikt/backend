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
type ReadyToPlay struct {
	Ready bool `json:"ready"`
}

// Example json-object: {"payloadtype":"AnswerFromPlayer","question":1,"votes":["aksel","alf"]}
type PayloadType struct {
	Type string `json:"payloadtype"`
}

type VotesToQuestion struct {
	Question int              `json:"question"`
	Votes     map[string]int `json:"votes"`
}

/* EXAMPLE JSON-obj
{
	"payloadtype": "VotesToQuestions",
	"votesToQuestions": [
		{
				"question": 1,
				"votes": {
						"aksel": 2,
						"alf": 10
				}
			}
		]
}
*/
type VotesToQuestions struct {
	PayloadType
	VotesToQuestions []VotesToQuestion `json:"votesToQuestions"`
}
