package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/selvinnsikt/backend/game"
	"github.com/selvinnsikt/backend/hub"
	"github.com/selvinnsikt/backend/model"
	"net/http"
)

// CreateRoom creates a new game room
func CreateHubHandler(w http.ResponseWriter, r *http.Request) {
	// Cors
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Creating a hub
	h, hubID := hub.NewHub()

	// Init the game
	go game.InitGame(h)

	// Return response to client with Hub ID
	json.NewEncoder(w).Encode(model.HubID{Hub: hubID})
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

	// Super shady origin check. Look under Origin Considerations here:
	// https://godoc.org/github.com/gorilla/websocket for documentation on
	// how to write a better one
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

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
