package api

import (
	"fmt"
	"net/http"

	"kwik-mq/internal/queue"
)

func health(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "OK")
}

func SetupRoutes() {
	http.HandleFunc("/health", health)

	http.HandleFunc("/queue/push", func(w http.ResponseWriter, r *http.Request) {
		CheckAccessToken(w, r, queue.QueuePushHandler)
	})
	http.HandleFunc("/queue/pop", func(w http.ResponseWriter, r *http.Request) {
		CheckAccessToken(w, r, queue.QueuePopHandler)
	})

	fmt.Println("Server is running...")
	http.ListenAndServe(":10526", nil)
}