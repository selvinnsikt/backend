package game

import (
	"../hub"
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
			g.Hub.BroadcastMsg(msg.Text)
		}
	}
}

// TODO: Create a switch which finds what type of message that was sent
// Define the different type of messages allowed
// For example:
// MsgType: 1
// Answer To A Question
func parseMessage(msg string) {
	switch msg {

	}
}
