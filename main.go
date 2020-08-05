package main

import (
	"./gameRoom"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	// Randomness
	rand.Seed(time.Now().UnixNano())

	//grs := gameRoom.NewGameRoom()
	//grs.Create("Aksel")

	gameRoom.InitGameRooms()

	log.Println("starting up server")
	log.Fatal(server())
}

func server() error{
	r := mux.NewRouter()

	r.HandleFunc("/join", gameRoom.JoinHandler).Methods("POST")

	r.HandleFunc("/create", gameRoom.CreateHandler).Methods("GET")

	return http.ListenAndServe(":8080",r)
}
