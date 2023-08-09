package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Mind-thatsall/pubsub/db"
	"github.com/Mind-thatsall/pubsub/models"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

func HandlerPostMessage(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postMessage(w, r, session)
	}
}

func postMessage(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var message models.Message

	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	message.MessageId = gocql.MustRandomUUID()

	q := session.Query(db.MessageTable.Insert()).BindStruct(message)
	if err := q.ExecRelease(); err != nil {
		log.Fatal(err)
	}

	users := getAllUsersFromChannel(message.ChannelId, session)

	broadcastMessage(users, message)
}

func getAllUsersFromChannel(channelId gocql.UUID, session *gocqlx.Session) []gocql.UUID {
	var users []gocql.UUID

	q := db.SubscriberTable.SelectBuilder("user_id").Where(qb.Contains("channels_id")).Query(*session).Bind(channelId)
	if err := q.Select(&users); err != nil {
		log.Fatal(err)
	}

	return users
}

func broadcastMessage(users []gocql.UUID, message models.Message) {
	for _, user_id := range users {
		if conn := Connections[user_id]; conn != nil {
			conn.WriteJSON(message)
		}
	}
}
