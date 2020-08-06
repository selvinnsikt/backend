package hub

import (
	"../model"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type GameHub interface {
	AddClientToHub(pc model.PlayerConnection)
	// Broadcast a message to all client connceted to the hub
	BroadcastMsg(msg string) error
	// Listen to all messages coming to the hub
	GetBroadcastChan() chan model.Message
}

var hubs *Hubs
var writeWait = 5 * time.Second

func InitHubs() {
	hubs = &Hubs{
		activeHubs: nil,
		RWMutex:    &sync.RWMutex{},
	}
}

type Hubs struct {
	// A slice of all the active hubs
	activeHubs []*Hub
	*sync.RWMutex
}

// One GameRoom
type Hub struct {
	hubID                  string
	clientsConn            map[string]*websocket.Conn
	addClientChan          chan *model.PlayerConnection
	removeClientChan       chan *model.PlayerConnection
	broadcastChan          chan model.Message
	numberClientsConnected int
	mutex                  *sync.RWMutex
}

// NewHub creates a new hub
func NewHub() (*Hub, string) {
	// Accessing global slice of hubs
	hubs.Lock()
	defer hubs.Unlock()
	hubID := generateHubID()

	log.Printf("creating a hub with ID: '%s'\n", hubID)
	// Creating a new hub
	h := &Hub{
		hubID:                  hubID,
		clientsConn:            make(map[string]*websocket.Conn),
		addClientChan:          make(chan *model.PlayerConnection),
		removeClientChan:       make(chan *model.PlayerConnection),
		broadcastChan:          make(chan model.Message),
		numberClientsConnected: 0,
		mutex:                  new(sync.RWMutex),
	}
	hubs.activeHubs = append(hubs.activeHubs, h)

	return h, h.hubID
}

func (h *Hub) run() {
	for {
		select {
		case pc := <-h.addClientChan:
			h.addClient(pc)
		case pc := <-h.removeClientChan:
			h.removeClient(pc)
		}
	}
}

func (h *Hub) removeClient(np *model.PlayerConnection) {
	defer np.Conn.Close()
	log.Printf("deleting '%s from hub '%s' with IP '%s'\n", np.Name, h.hubID, np.Conn.RemoteAddr().String())
	delete(h.clientsConn, np.Name)
	h.numberClientsConnected--
}

// readMessageFromClient reads incoming messages and sent it to incomingMsgChan
func (h *Hub) readMessageFromClient(pc *model.PlayerConnection) {
	for {
		var m model.Message
		err := pc.Conn.ReadJSON(&m)
		if err != nil {
			log.Println("Error: " + err.Error())
			h.removeClient(pc)
			return
		}
		h.broadcastChan <- m
	}
}

func (h *Hub) GetBroadcastChan() <-chan model.Message {
	return h.broadcastChan
}
func (h *Hub) BroadcastMsg(msg string) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for _, client := range h.clientsConn {
		client := client
		go func(c *websocket.Conn) {
			client.SetWriteDeadline(time.Now().Add(writeWait))
			err := client.WriteJSON(model.Message{Text: msg}) //{Type: -1, Msg: msg})
			if err != nil {
				log.Printf("error occurred while broadcasting message to IP '%s' , errorMsg: %s \n", client.RemoteAddr().String(), err.Error())
			}
		}(client)

	}
}

func (h *Hub) addClient(np *model.PlayerConnection) {
	// Add client to Game Room
	log.Printf("adding '%s to hub '%s' with IP '%s' \n", np.Name, h.hubID, np.Conn.RemoteAddr().String())
	h.clientsConn[np.Name] = np.Conn
	h.numberClientsConnected++

}

// addClientToHub adds the player to the given hub ID
func (h *Hub) AddClientToHub(pc model.PlayerConnection) {

	go h.run()

	// Adding the connection to gameroom
	h.addClientChan <- &pc

	// Read the messages sent from the client
	go h.readMessageFromClient(&pc)

}

// validateHubAndPlayerName validates parametes playerName and hub ID
func ValidateHubAndPlayerName(np model.NewPlayer) (*Hub, error) {
	// Get the room and check if the room exists
	h, err := getHub(np.HubID)
	if err != nil {
		return nil, err
	}

	// Check if name is available in given room
	if ok := h.playerNameAvailableInHub(np.Name); !ok {
		return nil, fmt.Errorf("name '%s' is already taken in hub '%s ", np.Name, np.HubID)
	}
	return h, nil
}

// generateHubID creates a 5 digit string and check if it is available
func generateHubID() string {
	var exists = false
	var roomID string
	for !exists {
		// Generate a random 5 digit number
		var n int
		for i := 0; i < 5; i++ {
			n = rand.Intn(9)
			roomID += strconv.Itoa(n)
		}
		// Check if room exist
		exists = hubExists(roomID)
	}
	return roomID
}

// getHub find the correct based on id and return pointer of room
func getHub(id string) (*Hub, error) {
	hubs.RLock()
	defer hubs.RUnlock()
	for _, gr := range hubs.activeHubs {
		if gr.hubID == id {
			return gr, nil
		}
	}
	return nil, fmt.Errorf("did not find any room with id '%s'", id)

}

// playerNameAvailableInHub loops through usernames and checks for duplicated names
func (h *Hub) playerNameAvailableInHub(n string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for name, _ := range h.clientsConn {
		if name == n {
			return false
		}
	}
	return true
}

func hubExists(id string) bool {
	for _, gr := range hubs.activeHubs {
		if gr.hubID == id {
			return false
		}
	}
	return true
}
