package main

import "log"

type sseClient struct {
	clientID string
	ch       chan string
}

type sseHub struct {
	register   chan *sseClient
	unregister chan *sseClient
	broadcast  chan string
}

var hub *sseHub

func newSSEClient(clientID string) *sseClient {
	// Buffered channel avoids blocking the hub on slow clients.
	return &sseClient{clientID: clientID, ch: make(chan string, 1)}
}

func newHub() *sseHub {
	// Central fan-out for HTML patches to all active SSE clients.
	return &sseHub{
		register:   make(chan *sseClient),
		unregister: make(chan *sseClient),
		broadcast:  make(chan string, 128),
	}
}

func (h *sseHub) start() {
	go func() {
		clients := make(map[*sseClient]struct{})
		for {
			select {
			case client := <-h.register:
				// Add new client to the active set.
				clients[client] = struct{}{}
				log.Printf("sse client register id=%s total=%d", client.clientID, len(clients))
			case client := <-h.unregister:
				// Remove client when their SSE connection closes.
				delete(clients, client)
				log.Printf("sse client unregister id=%s total=%d", client.clientID, len(clients))
			case patch := <-h.broadcast:
				// Best-effort broadcast; drop for slow clients to keep server responsive.
				for client := range clients {
					select {
					case client.ch <- patch:
					default:
						log.Printf("sse client slow, dropping patch")
					}
				}
			}
		}
	}()
}

func (h *sseHub) registerClient(client *sseClient) {
	// Add a client to the hub.
	h.register <- client
}

func (h *sseHub) unregisterClient(client *sseClient) {
	// Remove a client from the hub.
	h.unregister <- client
}

func (h *sseHub) broadcastPatch(patch string) {
	// Send a patch to all connected clients.
	log.Printf("sse broadcast size=%d", len(patch))
	h.broadcast <- patch
}
