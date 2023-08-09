package handlers

import (
	"fmt"
	"net/http"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	Connections     = make(map[gocql.UUID]*websocket.Conn)
	roomConnections = make(map[string][]*websocket.Conn)
)

func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection to webSocket", err)
	}

	ID := gocql.MustRandomUUID()
	Connections[ID] = conn
	initialMessage := map[string]string{
		"type":   "initial",
		"userId": ID.String(),
	}

	defer func() {
		conn.Close()
		delete(Connections, ID)
	}()

	conn.WriteJSON(initialMessage)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Printf("Received message: %s\n", msg)

	}
}
