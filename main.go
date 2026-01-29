package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/starfederation/datastar-go/datastar"
)

// data setup

type counterStore struct {
	value int64
}

func (c *counterStore) inc() {
	atomic.AddInt64(&c.value, 1)
}

func (c *counterStore) load() int64 {
	return atomic.LoadInt64(&c.value)
}

type incStore struct {
	mu  sync.Mutex
	ids []string
}

var counter counterStore
var incs incStore

func (s *incStore) push(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ids = append(s.ids, id)
}

func (s *incStore) load() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	ids := make([]string, len(s.ids))
	copy(ids, s.ids)
	return ids
}

const clientCookieName = "datastar-client-id"

var indexTpl = template.Must(template.ParseFiles("index.html"))

type mainViewData struct {
	ClientID string
	Counter  int64
	IncIDs   []string
}

// handlers

func index(w http.ResponseWriter, r *http.Request) {
	clientID := getOrSetClientID(w, r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	data := mainViewData{
		ClientID: clientID,
		Counter:  counter.load(),
		IncIDs:   []string{},
	}

	// index templaate (full html)
	err := indexTpl.ExecuteTemplate(w, "index", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func stream(w http.ResponseWriter, r *http.Request) {
	clientID := getOrSetClientID(w, r)
	sse := datastar.NewSSE(w, r)

	// loop with 10 fps
	for {
		if sse.IsClosed() {
			return
		}
		incIDs := incs.load()

		// main template, gets morphed by datastar
		var buf bytes.Buffer
		err := indexTpl.ExecuteTemplate(&buf, "main", mainViewData{
			ClientID: clientID,
			Counter:  counter.load(),
			IncIDs:   incIDs,
		})
		if err != nil {
			return
		}

		sse.PatchElements(buf.String())

		time.Sleep(100 * time.Millisecond)
	}
}

func increment(w http.ResponseWriter, r *http.Request) {
	clientID := getOrSetClientID(w, r)
	counter.inc()
	incs.push(clientID)
	w.WriteHeader(http.StatusNoContent)
}

// helpers

func getOrSetClientID(w http.ResponseWriter, r *http.Request) string {
	if cookie, err := r.Cookie(clientCookieName); err == nil && cookie.Value != "" {
		id := cookie.Value
		return id
	}

	id := newClientID()
	http.SetCookie(w, &http.Cookie{
		Name:     clientCookieName,
		Value:    id,
		Path:     "/",
		Expires:  time.Now().Add(1 * time.Hour),
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	})
	return id
}

func newClientID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

// main

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/stream", stream)
	http.HandleFunc("/increment", increment)
	http.ListenAndServe(":8080", nil)
}
