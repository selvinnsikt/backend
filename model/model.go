package model

import (
	"github.com/gorilla/websocket"
	"sync"
)

const (
	READY_TO_PLAY                     = "ReadyToPlay"
	FOUR_QUESTIONS                    = "FourQuestions"
	PLAYERS_VOTE_TO_QUESTION          = "PlayersVoteToQuestion"
	PLAYERS_VOTE_TO_QUESTION_DONE     = "PlayersVoteToQuestionDone"
	PLAYERS_VOTE_TO_QUESTION_RECIEVED = "PlayersVoteToQuestionReceived"
	PLAYERS_CONNECTED                 = "PlayersConnected"
	SELF_VOTE_ON_QUESTION             = "SelfVoteOnQuestion"
	SELF_VOTE_ON_QUESTION_RECEIVED    = "SelfVoteOnQuestionReceived"
	SELF_VOTE_ON_QUESTION_DONE        = "SelfVoteOnQuestionDone"
	MOST_VOTES                        = "mostVotes"
	NEUTRAL                           = "neutral"
	LEAST_VOTES                       = "leastVotes"
)
const (
	MAX_NUMBER_OF_ROUND = 4
	POINTS_MAX          = 3
	POINTS_NEUTRAL      = 1
	POINTS_ZERO         = 0
)

type HubID struct {
	Hub string `json:"hub"`
}

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

// Sent after the client is successfully connected with a websocket
type ConnSuccess struct {
	PayloadType
}
type PlayersConnected struct {
	PayloadType
	NumberConnected int `json:"numberConnected"`
}

type ReadyToPlay struct {
	PayloadType
	Ready  bool   `json:"ready"`
	Player string `json:"player,omitempty"`
}

type PayloadType struct {
	Type string `json:"payloadtype,omitempty"`
}

// Clients sends this to the server for voting on a question
type PlayersVotesToQuestion struct {
	PayloadType
	Question int            `json:"questionNumber"`
	Votes    map[string]int `json:"votes"`
}

// Server broadcasts this struct after received 'PlayersVotesToQuestion'
type PlayersVotesToQuestionReceived struct {
	PayloadType
	Question int    `json:"questionNumber"`
	Player   string `json:"player"`
}

type Questions struct {
	PayloadType
	Question []string `json:"questions"`
}
type Client struct {
	Conn  *websocket.Conn
	mutex *sync.RWMutex
}

type VotesToQuestions struct {
	PayloadType
	VotesToQuestions []PlayersVotesToQuestion `json:"votesToQuestions"`
}

type SelfVoteOnQuestion struct {
	PayloadType
	Question int    `json:"questionNumber"`
	Decision string `json:"decision"`
}
type SelfVoteOnQuestionReceived struct {
	PayloadType
	Question int    `json:"questionNumber"`
	Player   string `json:"player"`
}
type SelfVoteOnQuestionDone struct {
	PayloadType
	Question int `json:"questionNumber"`
	// map of playerName and how many points
	Points map[string]int `json:"points"`
}
