package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/selvinnsikt/backend/model"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

var NUMBER_OF_CLIENTS = 1

func main() {

		start := time.Now()

		startTest()

		used := time.Since(start)

		// program sleep for 1 second. Subtract out the sleep
		fmt.Printf("\n it used %s \n",used.String())


}

func startTest(){
	done := make(chan bool)
	numberRequest := make(chan string)

	msg := model.ReadyToPlay{
		PayloadType: model.PayloadType{Type: "ReadyToPlay"},
		Ready:       true,
	}
	fmt.Println("starting request loop")
	hubID := createGame()
	for i := 1; i <= NUMBER_OF_CLIENTS; i++ {
		go sendMsg(hubID, strconv.Itoa(i), msg, done, numberRequest)

	}

	fmt.Println("closing program")

	// Loop until all clients have received information
	var received = 0
	var p string
	players := make(map[string]int)
	for {
		p = <-numberRequest
		players[p] += 1
		received++
		//fmt.Println("Number received: " + strconv.Itoa(received))
		if players[p] == 5 {
			fmt.Printf("Player '%s' is done! \n", p)
		}
		if received == NUMBER_OF_CLIENTS*NUMBER_OF_CLIENTS {
			done <- true
			break
		}
	}
}

func createGame() string {
	u := url.URL{Scheme: "http", Host: "localhost:8080", Path: "/create"}
	res, err := http.Get(u.String())
	if err != nil {
		log.Fatal(err)
	}
	resRead, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	type Hub struct {
		HubID string `json:"hub"`
	}
	h := Hub{}
	err = json.Unmarshal(resRead, &h)
	if err != nil {
		log.Fatal(err)
	}

	return h.HubID

}

func sendMsg(hubID, player string, msg model.ReadyToPlay, done chan bool, numberRequest chan string) {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/join/" + hubID + "/" + player}
	log.Println("trying to dial with url " + u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Println(player + " connection problem")
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	err = conn.WriteJSON(msg)
	if err != nil {
		log.Println(player + " write problem")
		log.Fatal(err)
	}

	fmt.Printf("Player '%s' sent message sucessfully\n", player)
	go readMsg(conn, player, done, numberRequest)

}

func readMsg(c *websocket.Conn, player string, done chan bool, numberRequest chan string) {
	defer c.Close()
	var responseRead = 0
	for {
		_, m, err := c.ReadMessage()
		if string(m) == "" {
			continue
		}
		if err != nil {
			log.Println(player + " read problem")
			log.Fatal(err)
		}
		responseRead++
		numberRequest <- player
		if responseRead == NUMBER_OF_CLIENTS {
			<-done
			break
		}
	}
}
