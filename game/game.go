package game

import (
	"../hub"
	"../model"
	"encoding/json"
	"fmt"
	"log"
)

// Inits game and listen to a channel which received incoming messages from all
// clients who are connected to the hub

var game *Game

type Game struct {
	*hub.Hub
}

func InitGame(gh *hub.Hub) {
	game = new(Game)
	game.Hub = gh

	// Read messages from Hub
	go game.readHubMessages()
}

// readHubMessages reads all messages sent from the broadcast channel
func (g *Game) readHubMessages() {
	broadcastCh := g.GetBroadcastChan()
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
	case "AnswerFromPlayer":
		var a model.AnswerFromPlayer
		err = json.Unmarshal(d, &a)
		if err != nil {
			g.Hub.SendMsgToClient(fmt.Sprintf("unable to parse json-object '%s' to type '%s'", string(d), t), msg.Player)
			return
		}

		g.Hub.SendMsgToClient(a.Votes[0], msg.Player)
		//TODO: do something with the parsed message

	default:
		g.Hub.SendMsgToClient(fmt.Sprintf("'%s' is not of a valid message type", t), msg.Player)
		return
	}

}

// find what type of payload and the index the actual payload starts at
func getPayloadType(data []byte) (string, int, error) {
	var t string
	var p model.Payload

	// loop until the first comma (where the payloadtype in json-object ends)
	for i, d := range data {
		if string(d) == "," {
			t += "}"

			// validating the data
			err := json.Unmarshal([]byte(t), &p)
			if err != nil {
				return "", 0, err
			}
			if p.PayloadType == "" {
				return "", 0, fmt.Errorf("invalid json format")
			}
			// return if everything went fine
			i++
			return p.PayloadType, i, nil
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
