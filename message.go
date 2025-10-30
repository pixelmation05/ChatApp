package main
import (
 "time"
)

type message struct {
 Name    string    `json:"name"`    
 Message string    `json:"message"` 
 When    time.Time `json:"when"`    
}