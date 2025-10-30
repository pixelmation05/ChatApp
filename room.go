package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
)

type room struct {
	// forward is a channel that holds incoming messages
	// that should be forwarded to the other clients.
	forward chan *message

	join    chan *client
	leave   chan *client
	clients map[*client]bool
}
func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join: make(chan *client),
		leave: make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// handle client joining
			r.clients[client] = true
		case client := <- r.leave:
			delete(r.clients,client)
			close(client.send)
		case msg := <-r.forward:
			for client := range r.clients {
				client.send <-msg
			}
		}
	}
}
const (
	socketBufferSize =1024
	messageBufferSize = 256
)
var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}
func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    socket, err := upgrader.Upgrade(w, req, nil)
    if err != nil {
        log.Fatal("ServeHTTP:", err)
        return
    }
    
    // Get auth cookie and extract user data
    authCookie, err := req.Cookie("auth")
    if err != nil {
        log.Fatal("Failed to get auth cookie:", err)
        return
    }
    
    userData := objx.MustFromBase64(authCookie.Value)
    
    client := &client{
        socket:   socket,
        send:     make(chan *message, messageBufferSize),
        room:     r,
        userData: userData, // ✅ FIX: Set userData
    }
    
    r.join <- client
    defer func() {
        r.leave <- client
    }()
    go client.write()
    client.read()
}
// In your room.go broadcast method
