package gameRoom

import "../model"

type GameRoom interface {
	Create(np model.NewPlayer) error
	Join(np model.NewPlayer) error
}

// NewGameRoom creates a new GameRoom for controller to interact with gameRoom.go
func NewGameRoom() *GameRoom {

}

// Creates a new gameRoom
func (gr *gameRoom) Create(playerName string) (*gameRoom, error) {
	// Generate a string of six characters
	
	// Check if room exist

	// Create a new gameRoom
	return &gameRoom{
		GameRoomID: np.GameRoomID,
		Players:    []player{{Name: playerName}},
	}, nil
}
// Join will add a new player to an existing room
func (gr *gameRoom) Join(np model.NewPlayer) error {
	return nil
}

type gameRoom struct {
	GameRoomID string
	Players    []player
}

type player struct {
	Name string
}
