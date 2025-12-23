package main

import (
	"database/sql"
	"log"
	"os"
	"time"
	_ "github.com/lib/pq"
)

var db *sql.DB

type DBMessage struct { 
	Username  string
	Message   string
	CreatedAt time.Time
	RoomID    string
}

func initDB() {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Println("No DATABASE_URL")
		return
	}

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Printf("DB open error: %v", err)
		return
	}

	err = db.Ping()
	if err != nil {
		log.Printf("DB ping error: %v", err)
		db = nil
		return
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			username TEXT,
			message TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			room_id TEXT DEFAULT 'default'
		)
	`)
	if err != nil {
		log.Printf("Create table error: %v", err)
		db = nil
		return
	}

	log.Println("DB connected")
}

func saveMessage(username, message, roomID string) error {
	if db == nil {
		return nil
	}
	_, err := db.Exec(
		"INSERT INTO messages (username, message, room_id) VALUES ($1, $2, $3)",
		username, message, roomID,
	)
	return err
}

func getRecentMessages(roomID string, limit int) ([]DBMessage, error) { 
	if db == nil {
		return []DBMessage{}, nil 
	}
	rows, err := db.Query(
		"SELECT username, message, created_at, room_id FROM messages WHERE room_id = $1 ORDER BY created_at DESC LIMIT $2",
		roomID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []DBMessage 
	for rows.Next() {
		var m DBMessage 
		err := rows.Scan(&m.Username, &m.Message, &m.CreatedAt, &m.RoomID)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}

	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	return msgs, nil
}
