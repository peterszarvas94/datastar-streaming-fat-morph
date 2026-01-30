package main

import (
	"log"
	"net/http"
)

// main

func main() {
	// Initialize DB, start hub and background workers, then serve HTTP.
	hub = newHub()
	hub.start()
	log.Printf("server listening on http://localhost:8080")

	http.HandleFunc("/", index)
	http.HandleFunc("/stream", stream)
	http.HandleFunc("/increment", increment)
	http.HandleFunc("/decrement", decrement)
	http.HandleFunc("/reset", resetCounter)
	http.ListenAndServe(":8080", nil)
}
