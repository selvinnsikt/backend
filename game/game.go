package game

import (
	"encoding/json"
	"fmt"
	"github.com/selvinnsikt/backend/database"
	"github.com/selvinnsikt/backend/hub"
	"github.com/selvinnsikt/backend/model"
	"log"
	"sync"
	"time"
)

// Inits game and listen to a channel which received incoming messages from all
// clients who are connected to the hub

var game *Game

const (
	REQUIRED_VOTES_PER_QUESTIONS = 2
)

//
type Game struct {
	Hub *hub.Hub
	// Number of players that have sent ready
	NumberPlayersReady int
	// Interface to the database layer
	Database database.DB
	// Information about the active game
	ag activeGame
}

// ActiveGame manages information about the ongoing game
type activeGame struct {
	// Question that have been sent to the players
	questions []string
	// one activeGame consists of four rounds
	rounds []round
	mutex  *sync.RWMutex
}

type round struct {
	// Votes from players for current round
	playerVotes map[string]int
	// Self votes from players for current round
	// map[playerName]decision
	selfVotes map[string]string
}

func InitGame(h *hub.Hub) {
	// Init game struct
	game = new(Game)
	game.Hub = h
	game.Database = database.NewDatabase()

	game.ag.mutex = new(sync.RWMutex)

	// Read messages from Hub
	go game.readHubMessages()
}

// readHubMessages reads all messages sent from the broadcast channel
func (g *Game) readHubMessages() {
	broadcastCh := g.Hub.GetBroadcastChan()
	for {
		select {
		case msg := <-broadcastCh:
			log.Println("received message: " + msg.Text)
			g.handleDataFromHub(msg)
		}
	}
}

func (g *Game) handleDataFromHub(msg model.Message) {
	// get type of payload and index message starts
	t, i, err := getPayloadType([]byte(msg.Text))
	if err != nil {
		g.Hub.SendMsgToClient(fmt.Sprintf("unable to parse message: '%s'", msg.Text), msg.Player)
		return
	}
	// only get the wanted json-object
	d := removePayloadType(msg.Text, i)

	// find the correct struct
	switch t {
	case model.READY_TO_PLAY:
		// Parse the raw bytes to the correct struct
		var m model.ReadyToPlay
		err = json.Unmarshal(d, &m)
		if err != nil {
			g.Hub.SendMsgToClient(fmt.Sprintf("unable to parse json-object '%s' to type '%s' ; error: %s", string(d), t, err.Error()), msg.Player)
			return
		}
		// Add type to the message
		m.Type = model.READY_TO_PLAY

		// Add playername to the message
		m.Player = msg.Player

		// If player sent ready or not ready
		if m.Ready {
			g.NumberPlayersReady++
		} else {
			g.NumberPlayersReady--
		}

		// Broadcast to other players that the player is ready or not ready
		g.Hub.BroadcastMsg(m)

		// If everyone is ready
		if g.Hub.GetNumberOfClientsConnected() == g.NumberPlayersReady {
			// Starting the game
			g.beginGame()
		}
	case model.PLAYERS_VOTE_TO_QUESTION:
		// Read playerVotes from player
		var m model.PlayersVotesToQuestion
		err = json.Unmarshal(d, &m)
		if err != nil {
			g.Hub.SendMsgToClient(fmt.Sprintf("unable to parse json-object '%s' to type '%s' ; error: %s", string(d), t, err.Error()), msg.Player)
			return
		}

		// Check if the question number is valid
		// question number must be between 1-4
		if m.Question < 1 || m.Question > 4 {
			g.Hub.SendMsgToClient(fmt.Sprintf("%d is a invalid question number, must be between 1-4", m.Question), msg.Player)
			return
		}

		// If the sent playerVotes from client is valid
		if err := isValidNumberOfVotes(m.Votes); err != nil {
			g.Hub.SendMsgToClient(err.Error(), msg.Player)
			return
		}

		// Add playerVotes to game struct for this round
		g.ag.mutex.Lock()
		for player, votes := range m.Votes {
			// Question slice starts at index 0
			g.ag.rounds[m.Question-1].playerVotes[player] += votes
		}
		g.ag.mutex.Unlock()

		// Broadcast that a vote was received
		g.Hub.BroadcastMsg(model.PlayersVotesToQuestionReceived{
			PayloadType: model.PayloadType{Type: model.PLAYERS_VOTE_TO_QUESTION_RECIEVED},
			Question:    m.Question,
			Player:      msg.Player,
		})

		// Check if the round is done
		if m.Question == 4 {
			var totalVotes int

			// Counting total playerVotes for last question
			g.ag.mutex.RLock()
			for _, votes := range g.ag.rounds[model.MAX_NUMBER_OF_ROUND-1].playerVotes {
				totalVotes += votes
			}
			g.ag.mutex.RUnlock()

			// If everyone has voted this last round
			if totalVotes == REQUIRED_VOTES_PER_QUESTIONS*g.Hub.GetNumberOfClientsConnected() {
				// Remove this in production? Was needed during testing
				// to let the clients catch up with the last message 'PlayersVotesToQuestionReceived'
				time.Sleep(100 * time.Millisecond)

				// Signal the players that this stage is done
				g.Hub.BroadcastMsg(model.PayloadType{Type: model.PLAYERS_VOTE_TO_QUESTION_DONE})
			}
		}
	case model.SELF_VOTE_ON_QUESTION:
		var m model.SelfVoteOnQuestion
		err := json.Unmarshal(d, &m)
		if err != nil {
			g.Hub.SendMsgToClient(fmt.Sprintf("unable to parse json-object '%s' to type '%s' ; error: %s", string(d), t, err.Error()), msg.Player)
			return
		}
		// Check if the question number is valid
		// question number must be between 1-4
		if m.Question < 1 || m.Question > 4 {
			g.Hub.SendMsgToClient(fmt.Sprintf("%d is a invalid question number, must be between 1-4", m.Question), msg.Player)
			return
		}

		// Check if the Decision is a valid type
		if !(m.Decision == model.MOST_VOTES || m.Decision == model.NEUTRAL || m.Decision == model.LEAST_VOTES) {
			g.Hub.SendMsgToClient(fmt.Sprintf("%s is a invalid decision, must be 'mostVotes','neutral' or 'leastVotes'", m.Decision), msg.Player)
			return
		}

		// Register the self vote
		g.ag.mutex.Lock()
		g.ag.rounds[m.Question-1].selfVotes[msg.Player] = m.Decision
		g.ag.mutex.Unlock()

		// Respond to clients that a vote was registered
		g.Hub.BroadcastMsg(model.SelfVoteOnQuestionReceived{
			PayloadType: model.PayloadType{Type: model.SELF_VOTE_ON_QUESTION_RECEIVED},
			Question:    m.Question,
			Player:      msg.Player,
		})

		// If all the players have self-voted for this round
		if len(g.ag.rounds[m.Question-1].selfVotes) == g.Hub.GetNumberOfClientsConnected() {
			// Calculate points the different points
			responseMsg := model.SelfVoteOnQuestionDone{
				PayloadType: model.PayloadType{Type: model.SELF_VOTE_ON_QUESTION_DONE},
				Question:    m.Question,
				Points:      make(map[string]int),
			}
			// Get the current round
			r := g.ag.rounds[m.Question-1]

			// Find largest and smallest value
			max, min := maxAndMinVotes(r.playerVotes)

			// Give points to the players
			for p, decision := range r.selfVotes {
				v := r.playerVotes[p]
				// Most votes and self-vote is mostVotes
				if v == max && decision == model.MOST_VOTES {
					responseMsg.Points[p] = model.POINTS_MAX

					// Least Votes and self-vote is leastVotes
				} else if v == min && decision == model.LEAST_VOTES {
					responseMsg.Points[p] = model.POINTS_MAX

					// between max and min + self-vote is neutral
				} else if max > v && v > min && decision == model.NEUTRAL {
					responseMsg.Points[p] = model.POINTS_NEUTRAL

					// player missed on the self-vote
				} else {
					responseMsg.Points[p] = model.POINTS_ZERO
				}
			}

			// TODO: REMOVE IF EVERYTHING WORKS DURING PROD
			time.Sleep(100 * time.Millisecond)
			g.Hub.BroadcastMsg(responseMsg)
		}

	case model.PLAYERS_CONNECTED:
		g.Hub.SendMsgToClient(model.PlayersConnected{PayloadType: model.PayloadType{Type: model.PLAYERS_CONNECTED}, NumberConnected: g.Hub.GetNumberOfClientsConnected()}, msg.Player)
	default:
		g.Hub.SendMsgToClient(fmt.Sprintf("'%s' is not of a valid message type", t), msg.Player)
	}

}

// maxAndMinVotes find the max and min number of votes in the map
func maxAndMinVotes(votes map[string]int) (max, min int) {
	max = 0
	min = 9999 // some larger number
	for _, v := range votes {
		// If the number of votes is larger than max
		// then that is the new max
		if v > max {
			max = v
		}

		// If the number of votes is less than min
		// then that is the new min
		if min > v {
			min = v
		}
	}
	return max, min
}

// beginRound starts the round/game by sending the players four questions
func (g *Game) beginGame() {
	q, err := g.Database.GetQuestions()
	if err != nil {
		// TODO: broadcast error message
		// Implement a error
		//g.Hub.BroadcastMsg()
		return
	}

	// Init the game for four rounds
	for i := 0; i < model.MAX_NUMBER_OF_ROUND; i++ {
		g.ag.rounds = append(g.ag.rounds, round{
			playerVotes: make(map[string]int),
			selfVotes:   make(map[string]string),
		})
	}

	// TODO: Check that the questions have not already been sent in this hub
	// Recall this function if true

	// Append the questions to the slice of questions
	g.ag.questions = append(g.ag.questions, q...)

	// TODO: Remove this if nessecary. Used it to let the clients proccess the previous
	// message after reciving this
	time.Sleep(200 * time.Millisecond)

	// Send question to players
	g.Hub.BroadcastMsg(model.Questions{PayloadType: model.PayloadType{Type: model.FOUR_QUESTIONS}, Question: q})
}

// isValidNumberOfVotes validates that the sent playerVotes from a client is valid
func isValidNumberOfVotes(v map[string]int) error {
	var totalVotes int

	for _, num := range v {
		// min playerVotes is 1 and max is 2
		// Checks for negative numbers
		if !(num == 1 || num == 2) {
			return fmt.Errorf("vote on player per client is 1 or 2, not %d", num)
		}
		totalVotes += num
	}
	// players must vote 2 times
	if totalVotes != 2 {
		return fmt.Errorf("players must use 2 playerVotes, not %d", totalVotes)
	}

	return nil
}

// find what type of payload and the index the actual payload starts at
func getPayloadType(data []byte) (string, int, error) {
	var t string
	var p model.PayloadType

	// loop until the first comma (where the payloadtype in json-object ends)
	for i, d := range data {
		if string(d) == "," {
			t += "}"

			// validating the data
			err := json.Unmarshal([]byte(t), &p)
			if err != nil {
				return "", 0, err
			}
			if p.Type == "" {
				return "", 0, fmt.Errorf("invalid json format")
			}
			// return if everything went fine
			i++
			return p.Type, i, nil
		}
		t += string(d)
	}
	// this return should never happen
	return "", 0, fmt.Errorf("unable to parse the data")
}

// getPayload filters out the payloadtype from the data
func removePayloadType(d string, i int) []byte {
	p := "{"
	p += d[i:]
	return []byte(p)
}
