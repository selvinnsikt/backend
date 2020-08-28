# Explanation

## Run server with Docker

Build image:

    docker build -t selvinnsikt:**INSERT TAG** .

Run image:

    docker run -p 8080:8080 selvinnsikt:**INSERT TAG**

Test if image is running:<br>

    curl localhost:8080/create

Will respond with JSON-obj. Container will also log some information.

## Sequence diagrams

Website used for sequence diagrams: <https://sequencediagram.org/>

## Creating hub

![alt text](https://user-images.githubusercontent.com/20001253/91325122-1bd7fc00-e7c3-11ea-8a59-7c8e4af22d4f.png)

Sequence diagram code: <br>

    title creating hub/game
    client -> controller.go : GET /create
    controller.go -> hub.go: hub.NewHub()
    controller.go <-hub.go:return hub, hudID
    controller.go -> game.go: game.InitGame(h)
    game.go ->game.go: game.readHubMessages()
    controller.go ->client: response: {"Hub":"hubID"}

## Joining hub

![alt text](https://user-images.githubusercontent.com/20001253/91325126-1c709280-e7c3-11ea-88a6-7fdd86bb732a.png)
Sequence diagram code: <br>

    title joining hub/game
    client -> controller.go : GET /join/{hubID}/{playerName}
    controller.go -> hub.go: ValidateHubAndPlayerName(model.NewPlayer)
    hub.go ->controller.go: return (*Hub, err)
    controller.go ->client: upgrade connection to websocket
    controller.go -> hub.go: hub.AddClientToHub(model.PlayerConnection)

## Playing the game

![alt text](https://user-images.githubusercontent.com/20001253/91325130-1d092900-e7c3-11ea-8dfc-3cebc22692f0.png)
Sequence diagram code: <br>

```
title the game  (packets are sent over websocket conn)

group loop until received ready from all players
client -> game.go: {"payloadtype":"ReadyToPlay" , "ready":"true"}

opt player can press not ready button
client -> game.go: {"payloadtype":"ReadyToPlay" , "ready":"false"}
end

game.go ->client: {"payloadtype":"ReadyToPlay" , "ready":"true","player":"aksel"}

end

game.go -> db.go: getQuestions()
db.go ->game.go: return (array of four questions)

game.go ->client: {"payloadtype":"FourQuestions" , "questions": [ ] }

group loop until questionNumber > 4
group loop until players have answered one question

client -> game.go: {"payloadtype":"PlayersVoteToQuestion" ,"questionNumber": 1 ,  "players": ["alf","alf" ] }

game.go ->game.go: addVoteToQuestion()
game.go ->client:  {"payloadtype":"PlayersVoteToQuestionReceived" , "questionNumber": 1 , "player": "aksel"}
end

end
end

game.go -> client:{"payloadtype":"PlayersVoteToQuestionDone"}

group until questionNumber > 4
group until all players have voted on SelfVoteOnQuestion
client->game.go: {"payloadtype":"SelfVoteOnQuestion" , "questionNumber": 1, "decision":"mostVotes"}
note right of client: On 'SelfVoteOnQuestion' players can choose vote either 'mostVotes, leastVotes, neutral'
game.go ->client: {"payloadtype":"SelfVoteOnQuestion" , "questionNumber": 1, "player":"aksel"}
end
game.go ->client: {"payloadtype":"SelfVoteOnQuestionDone" , "questionNumber": 1, "players": {"aksel": 3, "alf": 0}
} }
end

```
