package main

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

var indexTpl = template.Must(template.ParseFiles("index.html"))

type mainViewData struct {
	ClientID string
	Counter  int64
	Actions  []counterAction
	ServerAt string
}

func index(w http.ResponseWriter, r *http.Request) {
	// Serve the initial page with the user's client ID.
	clientID := getOrSetClientID(w, r)
	log.Printf("index request client=%s", clientID)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	counterValue, actions := state.snapshot()
	data := mainViewData{
		ClientID: clientID,
		Counter:  counterValue,
		Actions:  actions,
		ServerAt: time.Now().Format(time.RFC3339),
	}

	err := indexTpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func stream(w http.ResponseWriter, r *http.Request) {
	// Open a long-lived SSE connection and wait for broadcast patches.
	clientID := getOrSetClientID(w, r)
	sse := datastar.NewSSE(w, r)
	client := newSSEClient(clientID)
	hub.registerClient(client)
	defer hub.unregisterClient(client)

	if patch, err := renderMainPatch(clientID); err == nil {
		log.Printf("sse initial patch client=%s size=%d", clientID, len(patch))
		sse.PatchElements(patch)
	} else {
		log.Printf("sse initial patch failed: client=%s err=%v", clientID, err)
	}

	for {
		select {
		case patch := <-client.ch:
			sse.PatchElements(patch)
		case <-r.Context().Done():
			return
		}
		if sse.IsClosed() {
			return
		}
	}
}

func increment(w http.ResponseWriter, r *http.Request) {
	// Increment counter and broadcast to all clients.
	handleCounterAction(w, r, actionIncrement)
}

func decrement(w http.ResponseWriter, r *http.Request) {
	// Decrement counter and broadcast to all clients.
	handleCounterAction(w, r, actionDecrement)
}

func resetCounter(w http.ResponseWriter, r *http.Request) {
	// Reset counter to zero and broadcast to all clients.
	handleCounterAction(w, r, actionReset)
}

func handleCounterAction(w http.ResponseWriter, r *http.Request, action counterActionType) {
	// Apply the action in memory, then broadcast HTML patch.
	clientID := getOrSetClientID(w, r)
	log.Printf("action request action=%s client=%s", action, clientID)
	counterValue := state.applyAction(action, clientID)
	log.Printf("counter action=%s client=%s value=%d", action, clientID, counterValue)
	if patch, err := renderMainPatch(clientID); err != nil {
		log.Printf("render patch failed: %v", err)
	} else {
		log.Printf("render patch ok: size=%d", len(patch))
		hub.broadcastPatch(patch)
	}
	w.WriteHeader(http.StatusNoContent)
}

func renderMainPatch(clientID string) (string, error) {
	// Render the patch for the main content area.
	counterValue, actions := state.snapshot()
	var buf bytes.Buffer
	err := indexTpl.ExecuteTemplate(&buf, "main", mainViewData{
		ClientID: clientID,
		Counter:  counterValue,
		Actions:  actions,
		ServerAt: time.Now().Format(time.RFC3339),
	})
	if err != nil {
		log.Printf("render main failed: client=%s err=%v", clientID, err)
		return "", err
	}
	return buf.String(), nil
}
