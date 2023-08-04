package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type Subscriber struct {
	ChannelId    gocql.UUID `db:"channel_id"`
	ServerId     gocql.UUID `db:"server_id"`
	SubscriberId gocql.UUID `db:"subscriber_id"`
}

type Message struct {
	MessageId gocql.UUID `db:"message_id"`
	ChannelId gocql.UUID `db:"channel_id"`
	ServerId  gocql.UUID `db:"server_id"`
	Content   string     `db:"content"`
	SenderId  gocql.UUID `db:"sender_id"`
	CreatedAt time.Time  `db:"created_at"`
}

type Server struct {
	ServerId gocql.UUID `db:"server_id"`
	Name     string     `db:"name"`
}

type Channel struct {
	ChannelId gocql.UUID `db:"channel_id"`
	Name      string     `db:"name"`
	ServerId  gocql.UUID `db:"server_id"`
}

var subscriberMetadata = table.Metadata{
	Name:    "subscribers",
	Columns: []string{"channel_id", "server_id", "subscriber_id"},
}

var serverMetadata = table.Metadata{
	Name:    "servers",
	Columns: []string{"server_id", "name"},
}

var channelMetadata = table.Metadata{
	Name:    "channels",
	Columns: []string{"channel_id", "name", "server_id"},
}

var messageMetadata = table.Metadata{
	Name:    "messages",
	Columns: []string{"message_id", "channel_id", "content", "created_at", "sender_id", "server_id"},
}

var (
	subscriberTable = table.New(subscriberMetadata)
	serverTable     = table.New(serverMetadata)
	channelTable    = table.New(channelMetadata)
	messageTable    = table.New(messageMetadata)
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	session := initScyllaDB()

	http.HandleFunc("/", handler)
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/api/fetch-servers", handlerFetchServers(session))
	http.HandleFunc("/api/fetch-channels", handlerFetchChannels(session))
	http.HandleFunc("/api/post-messages", handlerPostMessage(session))
	http.HandleFunc("/api/subscribe-channel", handlerSubToChannel(session))
	http.HandleFunc("/api/update-subscribe-channel", handlerUpdateSubToChannel(session))
	http.HandleFunc("/api/unsubscribe-channel", handlerUnSubToChannel(session))

	port := "8080"
	fmt.Printf("Server is running on http://localhost:%s\n", port)

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello you've accessed my Go server!")
}

var connections = make(map[gocql.UUID]*websocket.Conn)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading connection to webSocket", err)
	}

	ID := gocql.MustRandomUUID()
	connections[ID] = conn
	initialMessage := map[string]string{
		"type":   "initial",
		"userId": ID.String(),
	}

	defer func() {
		conn.Close()
		delete(connections, ID)
	}()

	conn.WriteJSON(initialMessage)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Error reading message:", err)
			break
		}

		fmt.Printf("Received message: %s", msg)

	}
}

func initScyllaDB() *gocqlx.Session {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "social"
	session, err := gocqlx.WrapSession(cluster.CreateSession())
	if err != nil {
		log.Fatal(err)
	}

	return &session
}

func handlerFetchServers(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fetchServers(w, r, session)
	}
}

func handlerFetchChannels(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fetchChannels(w, r, session)
	}
}

func handlerPostMessage(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postMessage(w, r, session)
	}
}

func handlerSubToChannel(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subToChannel(w, r, session)
	}
}

func handlerUpdateSubToChannel(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateSubToChannel(w, r, session)
	}
}

func handlerUnSubToChannel(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		unSubToChannel(w, r, session)
	}
}

func fetchServers(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var servers []Server

	q := session.Query(serverTable.SelectAll())
	if err := q.SelectRelease(&servers); err != nil {
		fmt.Println("here")
		log.Fatal(err)
	}

	data, err := json.Marshal(servers)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func fetchChannels(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var channels []Channel

	serverIdStr := r.URL.Query().Get("serverId")

	q := qb.Select("channels").Where(qb.EqLit("server_id", serverIdStr)).AllowFiltering().Query(*session)
	if err := q.Select(&channels); err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(channels)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func postMessage(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var message Message

	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	message.MessageId = gocql.MustRandomUUID()

	q := session.Query(messageTable.Insert()).BindStruct(message)
	if err := q.ExecRelease(); err != nil {
		log.Fatal(err)
	}
}

func subToChannel(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var subscriber Subscriber

	err := json.NewDecoder(r.Body).Decode(&subscriber)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	q := session.Query(subscriberTable.Insert()).BindStruct(subscriber)
	if err := q.ExecRelease(); err != nil {
		log.Fatal(err)
	}
}

func updateSubToChannel(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var subscriber Subscriber

	err := json.NewDecoder(r.Body).Decode(&subscriber)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	q := qb.Update("subscribers").Set("channel_id", "server_id").Where(qb.Eq("subscriber_id")).Query(*session).BindStruct(&subscriber)
	if err := q.ExecRelease(); err != nil {
		log.Fatal(err)
	}
}

func unSubToChannel(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var subscriber Subscriber

	err := json.NewDecoder(r.Body).Decode(&subscriber)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	q := qb.Delete("subscribers").Where(qb.Eq("subscriber_id")).Query(*session).BindStruct(&subscriber)
	if err := q.ExecRelease(); err != nil {
		log.Fatal(err)
	}
}
