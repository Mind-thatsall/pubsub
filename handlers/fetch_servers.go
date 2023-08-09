package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Mind-thatsall/pubsub/db"
	"github.com/Mind-thatsall/pubsub/models"
	"github.com/scylladb/gocqlx/v2"
)

func HandlerFetchServers(session *gocqlx.Session) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fetchServers(w, r, session)
	}
}

func fetchServers(w http.ResponseWriter, r *http.Request, session *gocqlx.Session) {
	var servers []models.Server

	q := session.Query(db.ServerTable.SelectAll())
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
