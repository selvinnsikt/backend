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
	"os"
	"sync"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	go func() {
		exitCode := m.Run()
		for _, p := range players {
			err := p.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("unable to correctly close the connection")
			}
			p.Conn.Close()
		}
		os.Exit(exitCode)
	}()

	// Start the server
	run()
}

var seqMutex sync.Mutex

// Ensures that these tests are run sequentially
func seq() func() {
	seqMutex.Lock()
	return func() {
		seqMutex.Unlock()
	}
}

var players []Connection
var playersName = []string{"aksel", "alf"}

type Connection struct {
	Conn  *websocket.Conn
	HubID string
}

func TestCreateAndJoinHub(t *testing.T) {
	defer seq()()

	res, err := http.Get("http://localhost:8080/create")
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("FAIL - execpeted status code %d, got '%d' ", http.StatusOK, res.StatusCode)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	var hubID model.HubID
	err = json.Unmarshal(b, &hubID)
	if err != nil {
		t.Errorf("FAIL - unable to unmarshal hubID bytes to struct")
	}
	// Joining the hub
	for i := 0; i < 2; i++ {
		conn, err := joinHub(hubID.Hub, playersName[i])
		if err != nil {
			t.Errorf("FAIL - unable to join the hub - %s ", err.Error())
		}
		// Setting the connection global
		players = append(players, Connection{Conn: conn, HubID: hubID.Hub})

	}

}

func joinHub(id, playerName string) (*websocket.Conn, error) {
	u := url.URL{Scheme: "ws", Host: "localhost:8080", Path: "/join/" + id + "/" + playerName}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	// Read the connection success message from server
	var successMsg model.ConnSuccess
	err = conn.ReadJSON(&successMsg)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func TestReadyToPlay(t *testing.T) {
	defer seq()()
	wg := sync.WaitGroup{}

	msg := model.ReadyToPlay{
		PayloadType: model.PayloadType{Type: model.READY_TO_PLAY},
		Ready:       true,
		Player:      "0",
	}

	// Send the two messages
	var err error
	for i, p := range players {
		msg.Player = playersName[i]
		err = p.Conn.WriteJSON(msg)
		if err != nil {
			t.Errorf("FAIL - unable to send ready msg to server - %s \n", err.Error())
		}
	}
	// Read the two messages
	for _, player := range players {
		wg.Add(1)
		go readFromConn(player, t, &wg)
	}
	wg.Wait()

}

func readFromConn(player Connection, t *testing.T, wg *sync.WaitGroup) {
	var numberReceived int
	var msgReceive model.ReadyToPlay

	for {
		err := player.Conn.ReadJSON(&msgReceive)
		if err != nil {
			t.Errorf("FAIL - unable to read message from server - %s \n", err.Error())
		}
		// Check if the correct type was sent
		if msgReceive.PayloadType.Type != model.READY_TO_PLAY {
			t.Errorf("FAIL - expected %s, got %s", model.READY_TO_PLAY, msgReceive.PayloadType.Type)
		}

		// Check if they sent true
		if msgReceive.Ready != true {
			t.Errorf("FAIL -  expected true from player %s, got %t", msgReceive.Player, msgReceive.Ready)
		}

		numberReceived++
		// I have read responses from all the players
		if numberReceived == len(players) {
			wg.Done()
			return
		}
	}
}

func TestReceivingQuestions(t *testing.T) {
	defer seq()()
	wg := sync.WaitGroup{}

	for i, p := range players {
		wg.Add(1)

		go func(num int, player Connection) {
			var receiveMsg model.Questions
			for {
				err := player.Conn.ReadJSON(&receiveMsg)
				if err != nil {
					t.Errorf("FAIL - error reading json-object from server - %s", err.Error())
				}
				if receiveMsg.PayloadType.Type != model.FOUR_QUESTIONS {
					t.Errorf("FAIL - expected %s, got %s", model.FOUR_QUESTIONS, receiveMsg.PayloadType.Type)
				}
				wg.Done()
				return
			}

		}(i, p)
	}
	wg.Wait()
}

func TestAnswerToQuestions(t *testing.T) {
	defer seq()()

	wgVotesToQuestions := new(sync.WaitGroup)

	// READING THE ANSWERS
	wgVotesToQuestions.Add(2)
	for _, player := range players {
		go func(p Connection) {
			var counter int
			for {
				// Starting to read
				var msgRes model.PlayersVotesToQuestionReceived
				err := p.Conn.ReadJSON(&msgRes)
				if err != nil {
					t.Errorf("ERROR - unable to read msg from server - %s", err.Error())
				}

				// First message is read
				counter += msgRes.Question
				// Return after all messages are received
				if counter == 10*len(players) { // 1+2+3+4 = 10 (the questions expected to be received)
					wgVotesToQuestions.Done()
					break
				}
			}

		}(player)
	}

	// Loup for every question (four times)
	for i := 0; i < 4; i++ {
		// Send answer to server
		msg := model.PlayersVotesToQuestion{
			PayloadType: model.PayloadType{Type: model.PLAYERS_VOTE_TO_QUESTION},
			Question:    i + 1,
			Votes:       map[string]int{"aksel": 2},
		}
		for _, p := range players {
			// Writing to server
			err := p.Conn.WriteJSON(&msg)
			if err != nil {
				t.Errorf("ERROR - unable to send msg to server - %s", err.Error())
			}
		}
	}

	// Wait for the 'ReceivedAnswerToQuestion' reads
	wgVotesToQuestions.Wait()
	for _, p := range players {
		var msgRec model.PayloadType
		err := p.Conn.ReadJSON(&msgRec)
		if err != nil {
			t.Errorf("ERROR -  unable to read msg from server - %s", err.Error())
		}
		if msgRec.Type != model.PLAYERS_VOTE_TO_QUESTION_DONE {
			t.Errorf("FAIL - expected %s, got %s ", model.PLAYERS_VOTE_TO_QUESTION_DONE, msgRec.Type)
		}

	}
}

func TestSelfVoteOnQuestion(t *testing.T) {
	defer seq()()

	wgWaitForRead := new(sync.WaitGroup)

	// Read from the server
	for _, player := range players {
		wgWaitForRead.Add(1)
		go func(p Connection) {
			var msgRec model.SelfVoteOnQuestionReceived
			var msgDone model.SelfVoteOnQuestionDone

			// Expecting 12 messages
			// Players  *  Rounds  +  msgDone
			//   2  	*    4     +    4
			maxLoop := len(players)*model.MAX_NUMBER_OF_ROUND + model.MAX_NUMBER_OF_ROUND
			for i := 1; i <= maxLoop; i++ {

				// if statement works only for two players
				if i%3 == 0 {
					err := p.Conn.ReadJSON(&msgDone)
					if err != nil {
						t.Errorf("ERROR -  unable to read msg from server - %s", err.Error())
					}
					if msgDone.PayloadType.Type != model.SELF_VOTE_ON_QUESTION_DONE {
						t.Errorf("FAIL - invalid json object from server - expected %s, got %s", model.SELF_VOTE_ON_QUESTION_DONE, msgDone.PayloadType.Type)
					}
					fmt.Println(msgDone)
				} else {
					err := p.Conn.ReadJSON(&msgRec)
					if err != nil {
						t.Errorf("ERROR -  unable to read msg from server - %s", err.Error())
					}
					if msgRec.PayloadType.Type != model.SELF_VOTE_ON_QUESTION_RECEIVED {
						t.Errorf("FAIL  - invalid json object from server - expected %s, got %s", model.SELF_VOTE_ON_QUESTION_RECEIVED, msgRec.PayloadType.Type)
					}
				}
			}
			wgWaitForRead.Done()
		}(player)
	}

	// Write to the server
	for i := 0; i < 4; i++ {
		msg := model.SelfVoteOnQuestion{
			PayloadType: model.PayloadType{Type: model.SELF_VOTE_ON_QUESTION},
			Question:    i + 1,
			Decision:    model.MOST_VOTES,
		}
		for _, p := range players {
			err := p.Conn.WriteJSON(&msg)
			if err != nil {
				t.Errorf("FAIL - unable to write to server - %s", err.Error())
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	wgWaitForRead.Wait()
}
