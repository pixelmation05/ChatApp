package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type client struct {
	socket   *websocket.Conn
	send     chan *message
	room     *room
	userData map[string]any
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		var msg *message
		err := c.socket.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}
		msg.When = time.Now()
		
		// ✅ FIX: Safe nil checking for userData
		if c.userData != nil {
			if name, ok := c.userData["name"].(string); ok && name != "" {
				msg.Name = name
			} else {
				msg.Name = "Anonymous"
			}
		} else {
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
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}