package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Mind-thatsall/pubsub/db"
	"github.com/Mind-thatsall/pubsub/models"
	"github.com/scylladb/gocqlx/v2"
)

func HandlerSubToChannel(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subToChannel(w, r, session)
	}
}

func subToChannel(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var subscriber models.Subscriber

	err := json.NewDecoder(r.Body).Decode(&subscriber)
	if err != nil {
		http.Error(w, "Failed to parse request body", http.StatusBadRequest)
		return
	}

	q := session.Query(db.SubscriberTable.Insert()).BindStruct(subscriber)
	if err := q.ExecRelease(); err != nil {
		log.Fatal(err)
	}
}
