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
}

func index(w http.ResponseWriter, r *http.Request) {
	clientID := getOrSetClientID(w, r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := mainViewData{
		ClientID: clientID,
		Counter:  counter.load(),
		Actions:  counterEvents.load(),
	}

	err := indexTpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func stream(w http.ResponseWriter, r *http.Request) {
	clientID := getOrSetClientID(w, r)
	sse := datastar.NewSSE(w, r)

	for {
		if sse.IsClosed() {
			return
		}
		actions := counterEvents.load()

		var buf bytes.Buffer
		err := indexTpl.ExecuteTemplate(&buf, "main", mainViewData{
			ClientID: clientID,
			Counter:  counter.load(),
			Actions:  actions,
		})
		if err != nil {
			return
		}

		sse.PatchElements(buf.String())
		time.Sleep(100 * time.Millisecond)
	}
}

func increment(w http.ResponseWriter, r *http.Request) {
	handleCounterAction(w, r, actionIncrement)
}

func decrement(w http.ResponseWriter, r *http.Request) {
	handleCounterAction(w, r, actionDecrement)
}

func resetCounter(w http.ResponseWriter, r *http.Request) {
	handleCounterAction(w, r, actionReset)
}

func handleCounterAction(w http.ResponseWriter, r *http.Request, action counterActionType) {
	clientID := getOrSetClientID(w, r)
	applyCounterAction(action)
	counterEvents.push(counterAction{ClientID: clientID, Action: action})
	log.Printf("counter action=%s client=%s value=%d", action, clientID, counter.load())
	enqueueCounterEvent(counterEvent{
		ClientID:  clientID,
		Action:    action,
		CreatedAt: time.Now().UTC(),
	})
	w.WriteHeader(http.StatusNoContent)
}
