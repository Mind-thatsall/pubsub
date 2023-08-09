package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Mind-thatsall/pubsub/db"
	"github.com/Mind-thatsall/pubsub/models"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

func HandlerFetchChannels(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fetchChannels(w, r, session)
	}
}

func fetchChannels(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var channels []models.Channel

	serverIdStr := r.URL.Query().Get("serverId")

	q := db.ChannelTable.SelectBuilder().Where(qb.EqLit("server_id", serverIdStr)).AllowFiltering().Query(*session)
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
