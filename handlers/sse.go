package handlers

import (
	"io"
	"sync"

	"github.com/gin-gonic/gin"
)

var clients = make(map[chan string]bool)
var mutex = &sync.Mutex{}

// Broadcast sends a message to all connected SSE clients
func Broadcast(message string) {
	mutex.Lock()
	defer mutex.Unlock()
	for client := range clients {
		// Non-blocking send to avoid getting stuck on a dead client
		select {
		case client <- message:
		default:
		}
	}
}

// SSEHandler handles the Server-Sent Events endpoint
func SSEHandler(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// Flush headers immediately
	c.Writer.Flush()

	clientChan := make(chan string, 10)
	mutex.Lock()
	clients[clientChan] = true
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		delete(clients, clientChan)
		close(clientChan)
		mutex.Unlock()
	}()

	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-clientChan:
			if !ok {
				return false
			}
			c.SSEvent("message", msg)
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})
}
