package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
)

type Message struct {
	Text string `json:"text"`
}

var gameroomID = flag.String("id", "", "enter a game room ID")
var playerName = flag.String("player", "", "enter a player name")

func main() {
	flag.Parse()
	if *gameroomID == "" {
		fmt.Println("game room ID is empty..")
		return
	}
	if *playerName == "" {
		fmt.Println("player name is empty")
		return
	}

	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/join/" + *gameroomID + "/" + *playerName}
	log.Println("trying to dial with url " + u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// receive
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("Error receiving message: ", err.Error())
				break
			}
			fmt.Println("Message: ", string(msg))
		}
	}()

	// send
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		err := conn.WriteMessage(1, []byte(text))
		if err != nil {
			fmt.Println("Error sending message: ", err.Error())
			break
		}
	}
}
