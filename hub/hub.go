package hub

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/selvinnsikt/backend/model"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type GameHub interface {
	AddClientToHub(pc model.PlayerConnection)
	// Broadcast a message to all client connceted to the hub
	BroadcastMsg(msg interface{}) error
	// Listen to all messages coming to the hub
	GetBroadcastChan() chan model.Message
	// sends a message to the client
	SendMsgToClient(msg interface{}, player string) error
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
	clientsConn            map[string]Client
	addClientChan          chan *model.PlayerConnection
	removeClientChan       chan *model.PlayerConnection
	broadcastChan          chan model.Message
	numberClientsConnected int
	mutex                  *sync.RWMutex
}

type Client struct {
	Conn  *websocket.Conn
	mutex *sync.RWMutex
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
		clientsConn:            make(map[string]Client),
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
func (h *Hub) addClient(np *model.PlayerConnection) {
	// Add client to Game Room
	log.Printf("adding '%s to hub '%s' with IP '%s' \n", np.Name, h.hubID, np.Conn.RemoteAddr().String())
	h.clientsConn[np.Name] = Client{
		Conn:  np.Conn,
		mutex: new(sync.RWMutex),
	}
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

// readMessageFromClient reads incoming messages and sent it to incomingMsgChan
func (h *Hub) readMessageFromClient(pc *model.PlayerConnection) {
	var m model.Message
	for {
		_, msg, err := pc.Conn.ReadMessage()
		if err != nil {
			log.Println("Error: " + err.Error())
			h.removeClient(pc)
			return
		}
		// add name of client who sent the message
		m.Player = pc.Name
		m.Text = string(msg)
		h.broadcastChan <- m
	}
}
func (h *Hub) SendMsgToClient(msg interface{}, player string) {
	if c, ok := h.clientsConn[player]; ok {
		err := c.sendMsg(msg)
		if err != nil {
			h.removeClient(&model.PlayerConnection{
				Name: player,
				Conn: c.Conn,
			})
		}
	} else {
		log.Printf("did not find any player in hub '%s' with name '%s'\n", h.hubID, player)
	}
}

func (c *Client) sendMsg(msg interface{}) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		log.Println(err)
		return err
	}
	err = c.Conn.WriteJSON(msg)
	if err != nil {
		return err
	}
	return nil
}

func (h *Hub) GetBroadcastChan() <-chan model.Message {
	return h.broadcastChan
}
func (h *Hub) BroadcastMsg(msg interface{}) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	for _, client := range h.clientsConn {
		client := client
		go func(c *Client) {
			err := c.sendMsg(msg)
			if err != nil {
				log.Printf("error occurred while broadcasting message to IP '%s' , errorMsg: %s \n", c.Conn.RemoteAddr().String(), err.Error())
			}
		}(&client)

	}
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
