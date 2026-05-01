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

	http.HandleFunc("/queue/push", queue.QueuePush)
	http.HandleFunc("/queue/pop", queue.QueuePop)

	fmt.Println("Server is running...")
	http.ListenAndServe(":10526", nil)
}