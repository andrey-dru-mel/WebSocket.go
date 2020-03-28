package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"encoding/json"
	"net/http"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan []byte)           // broadcast channel

type number struct {
	Type string `json:"type"`
	Number int `json:"number"`
}

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	// upgrade this connection to a WebSocket
	// connection
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	log.Println("Client Connected")
	if err != nil {
		log.Println(err)
	}

	numberOfClients()
	// listen indefinitely for new messages coming
	// through on our WebSocket connection
	reader(ws)
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			delete(clients, conn)
			numberOfClients()
			return
		}
		// print out that message for clarity
		fmt.Println(string(p))

		broadcast <- p

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}

func setupRoutes() {
	http.HandleFunc("/ws", wsEndpoint)
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Println(err)
				client.Close()
				delete(clients, client)
				numberOfClients()
			}
		}
	}
}

func numberOfClients(){
	m := number{Type: "number", Number: len(clients)}
	msg, err := json.Marshal(m)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	for client := range clients {
		err := client.WriteMessage(1, msg)
		if err != nil {
			log.Println(err)
			client.Close()
			delete(clients, client)
			numberOfClients()
		}
	}
}

func main() {
	fmt.Println("Start")
	go handleMessages()
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8081", nil))
}