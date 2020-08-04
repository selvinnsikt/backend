package model

// NewPlayer is used by both /newGameRoom and /joinGameRoom
type NewPlayer struct {
	Name       string `json:"name"`
	GameRoomID string `json:"game_room_id"`
}
