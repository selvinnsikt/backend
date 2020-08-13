package main

import (
	"./controller"
	"./hub"
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	// Randomness
	rand.Seed(time.Now().UnixNano())

	hub.InitHubs()

	log.Println("starting up server")
	log.Fatal(server())

}

func server() error {

	r := mux.NewRouter()

	r.HandleFunc("/join/{hub}/{player}", controller.JoinRoomHandler)
	r.HandleFunc("/create", controller.CreateHubHandler)

	return http.ListenAndServe(":8080", r)
}
