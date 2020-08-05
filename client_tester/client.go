package main

import (
	"../model"
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Message struct {
	Text string `json:"text"`
}

func main() {

	c := http.Client{}

	np := model.NewPlayer{
		Name:       "aksel",
		GameRoomID: "07608",
	}
	b, _ := json.Marshal(&np)

	req, err := http.NewRequest("POST", "http://localhost:8080/join", bytes.NewReader(b))
	if err != nil {
		log.Fatal(err)
	}

	res, err := c.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: creating a game room is working. Next is joining a game room.
	// Do I have to dial or can I figure out a way to upgrade the connection?
	websocket.DefaultDialer.Dial()

	// receive
	var m Message
	go func() {
		for {
			err := websocket.JSON.Receive(ws, &m)
			if err != nil {
				fmt.Println("Error receiving message: ", err.Error())
				break
			}
			fmt.Println("Message: ", m)
		}
	}()

	// send
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		m := Message{
			Text: text,
		}
		err = websocket.JSON.Send(ws, m)
		if err != nil {
			fmt.Println("Error sending message: ", err.Error())
			break
		}
	}
}
