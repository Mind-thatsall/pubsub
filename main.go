package main

import (
	"fmt"
	"net/http"

	"github.com/Mind-thatsall/pubsub/db"
	"github.com/Mind-thatsall/pubsub/handlers"
	"github.com/Mind-thatsall/pubsub/middleware"
)

func main() {
	session := db.InitScyllaDB()

	mux := http.NewServeMux()

	mux.HandleFunc("/", handler)
	mux.HandleFunc("/api/fetch-servers", handlers.HandlerFetchServers(session))
	mux.HandleFunc("/api/fetch-channels", handlers.HandlerFetchChannels(session))
	mux.HandleFunc("/api/post-messages", handlers.HandlerPostMessage(session))
	mux.HandleFunc("/api/subscribe-channel", handlers.HandlerSubToChannel(session))
	mux.HandleFunc("/api/unsubscribe-channel", handlers.HandlerUnSubToChannel(session))

	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/ws", handlers.WsHandler)

	fmt.Println("here")
	handlerMiddleware := middleware.InitCorsMiddleware(mux)
	go func() {
		errApi := http.ListenAndServe(":8080", handlerMiddleware)
		if errApi != nil {
			fmt.Print(errApi)
		}
	}()

	errWS := http.ListenAndServe(":8081", wsMux)
	if errWS != nil {
		fmt.Print(errWS)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello you've accessed my Go server!")
}
