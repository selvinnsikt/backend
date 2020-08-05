package gameRoom

import (
	"../model"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
)

var rooms *gameRooms

func InitGameRooms() {
	rooms = &gameRooms{
		roomSlice: nil,
		RWMutex:   &sync.RWMutex{},
	}
}

type gameRooms struct {
	// A slice of all the active rooms
	roomSlice []*gameRoom
	*sync.RWMutex
}

// One GameRoom
type gameRoom struct {
	gameRoomID       string
	clientsConn      map[string]*websocket.Conn
	clientsName      map[string]string
	addClientChan    chan *websocket.Conn
	removeClientChan chan *websocket.Conn
	broadcastChan    chan model.Message
}

func NewGameRoom(name string) *gameRoom {
	log.Printf("creating a new game room with ID: '%s'\n",name)
	return &gameRoom{
		gameRoomID:       name,
		clientsConn:      make(map[string]*websocket.Conn),
		clientsName:      make(map[string]string),
		addClientChan:    make(chan *websocket.Conn),
		removeClientChan: make(chan *websocket.Conn),
		broadcastChan:    make(chan model.Message),
	}
}

func (gr *gameRoom) Run() {
	log.Println("running function Run()")
	for {
		select {
		case conn := <-gr.addClientChan:
			gr.addClient(conn)
		case conn := <-gr.removeClientChan:
			gr.removeClient(conn)
		case a := <-gr.broadcastChan:
			gr.broadcastMessage(a)
		}
	}
}

func (gr *gameRoom) removeClient(conn *websocket.Conn) {
	delete(gr.clientsConn, conn.LocalAddr().String())
}

func (gr *gameRoom) addClient(conn *websocket.Conn) {
	// Add client to Game Room
	log.Printf("adding new client to GameRoom from IP '%s' \n", conn.LocalAddr().String())
	gr.clientsConn[conn.RemoteAddr().String()] = conn

}

func (gr *gameRoom) broadcastMessage(m model.Message) {
	for _, conn := range gr.clientsConn {
		err := conn.WriteJSON(m)
		if err != nil {
			log.Println("Error broadcasting message: ", err)
			return
		}
	}
}
func addclientToRoom(ws *websocket.Conn, gr *gameRoom, playerName string) {
	go gr.Run()

	// Adding the connection to gameroom
	gr.addClientChan <- ws

	for {
		var a model.Answer
		err := 	ws.WriteJSON(&a)
		if err != nil {
			gr.broadcastChan <- model.Message{Text: err.Error()}
			gr.removeClient(ws)
			return
		}
		// Broadcast to other channels that a answered was received
		gr.broadcastChan <- model.Message{Text: "answer++"}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func JoinHandler(w http.ResponseWriter, r *http.Request) {
	// Parsing the request
	var np model.NewPlayer
	if err := json.NewDecoder(r.Body).Decode(&np); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the room and check if the room exists
	room, err := getRoom(np.GameRoomID)
	if err != nil {
		http.Error(w, fmt.Sprintf("did not find any room with id '%s'", np.GameRoomID), http.StatusBadRequest)
		return
	}

	// Upgrades connection from HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// Add player to room
	addclientToRoom(conn, room, np.Name)
}

// CreateHandler creates a new game room
func CreateHandler(w http.ResponseWriter, r *http.Request) {
	rooms.Lock()
	defer rooms.Unlock()

	roomID := generateRoomID()
	gameRoom := NewGameRoom(roomID)
	rooms.roomSlice = append(rooms.roomSlice, gameRoom)

	// Return response to client with gameRoom ID
	json.NewEncoder(w).Encode(map[string]string{"gameRoom": roomID})
}

// generateRoomID creates a 5 digit string and check if it is available
func generateRoomID() string {
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
		exists = roomExists(roomID)
	}
	return roomID
}

// getRoom find the correct based on id and return pointer of room
func getRoom(id string) (*gameRoom, error) {
	rooms.RLock()
	defer rooms.RUnlock()
	for _, gr := range rooms.roomSlice {
		if gr.gameRoomID == id {
			return gr, nil
		}
	}
	return nil, fmt.Errorf("did not find any room with id '%s'", id)

}

func roomExists(id string) bool {
	for _, gr := range rooms.roomSlice {
		if gr.gameRoomID == id {
			return false
		}
	}
	return true
}
