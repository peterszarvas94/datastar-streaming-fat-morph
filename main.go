package main

import (
	"log"
	"net/http"
)

// main

func main() {
	var err error
	db, err = initDB(sqlitePath)
	if err != nil {
		log.Fatalf("failed to init sqlite: %v", err)
	}
	eventQueue = make(chan counterEvent, eventQueueSize)
	startCounterEventWorker(db, eventQueue)
	log.Printf("server listening on http://localhost:8080")

	http.HandleFunc("/", index)
	http.HandleFunc("/stream", stream)
	http.HandleFunc("/increment", increment)
	http.HandleFunc("/decrement", decrement)
	http.HandleFunc("/reset", resetCounter)
	http.ListenAndServe(":8080", nil)
}
