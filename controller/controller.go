package controller

import (
	"../game"
	"../hub"
	"../model"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"net/http"
)

// CreateRoom creates a new game room
func CreateHubHandler(w http.ResponseWriter, r *http.Request) {
	// Creating a hub
	h, hubID := hub.NewHub()

	// Init the game
	go game.InitGame(h)

	// Return response to client with Hub ID
	json.NewEncoder(w).Encode(map[string]string{"Hub": hubID})
}

var upgrader = websocket.Upgrader{}

func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	// Parsing the request
	vars := mux.Vars(r)
	np := model.NewPlayer{
		Name:  vars["player"],
		HubID: vars["hub"],
	}
	if np.Name == "" {
		http.Error(w, "player name in url is empty", http.StatusBadRequest)
		return
	}
	if np.HubID == "" {
		http.Error(w, "hub ID in url is empty", http.StatusBadRequest)
		return
	}

	// Trying to join the room
	h, err := hub.ValidateHubAndPlayerName(np)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Upgrades connection from HTTP to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.AddClientToHub(model.PlayerConnection{
		Name: np.Name,
		Conn: conn,
	})

}
