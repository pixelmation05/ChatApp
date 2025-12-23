package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/objx"
)


type message struct {
	Name    string    `json:"name"`
	Message string    `json:"message"`
	When    time.Time `json:"when"`
}

// Client type
type client struct {
	socket   *websocket.Conn
	send     chan *message
	room     *room
	userData map[string]interface{}
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			log.Printf("read error: %v", err)
			return
		}
		msg.When = time.Now()

		if c.userData != nil {
			if name, ok := c.userData["name"].(string); ok && name != "" {
				msg.Name = name
			}
		}
		if msg.Name == "" {
			msg.Name = "Anonymous"
		}

		c.room.forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteJSON(msg)
		if err != nil {
			log.Printf("write error: %v", err)
			break
		}
	}
}

// Room type
type room struct {
	forward chan *message
	join    chan *client
	leave   chan *client
	clients map[*client]bool
}

func newRoom() *room {
	return &room{
		forward: make(chan *message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *room) run() {
	for {
		select {
		case cl := <-r.join: // CHANGED: 'client' to 'cl'
			r.clients[cl] = true

		
			go func(c *client) {
				msgs, err := getRecentMessages("default", 50)
				if err == nil {
					for _, msg := range msgs {
						chatMsg := &message{
							Name:    msg.Username,
							Message: msg.Message,
							When:    msg.CreatedAt,
						}
						select {
						case c.send <- chatMsg:
						default:
						}
					}
				}
			}(cl) // CHANGED: Pass 'cl' instead of 'client'

		case cl := <-r.leave: // CHANGED: 'client' to 'cl'
			delete(r.clients, cl)
			close(cl.send)

		case msg := <-r.forward:
		
			go func() {
				if msg.Name != "" && msg.Message != "" {
					saveMessage(msg.Name, msg.Message, "default")
				}
			}()

			
			for cl := range r.clients {
				select {
				case cl.send <- msg:
				default:
					delete(r.clients, cl)
					close(cl.send)
				}
			}
		}
	}
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Printf("upgrade error: %v", err)
		return
	}

	authCookie, err := req.Cookie("auth")
	if err != nil {
		socket.Close()
		return
	}

	userData := objx.MustFromBase64(authCookie.Value)

	cl := &client{ // CHANGED: 'client' to 'cl'
		socket:   socket,
		send:     make(chan *message, 256),
		room:     r,
		userData: userData,
	}

	r.join <- cl // CHANGED: 'client' to 'cl'
	defer func() { r.leave <- cl }() // CHANGED: 'client' to 'cl'
	go cl.write() // CHANGED: 'client' to 'cl'
	cl.read() // CHANGED: 'client' to 'cl'
}
