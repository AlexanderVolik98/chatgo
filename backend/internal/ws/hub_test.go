package ws

import (
	"sync"
	"testing"
)

// TestHubConcurrentAccess stresses the hub from many goroutines; run with -race.
func TestHubConcurrentAccess(t *testing.T) {
	hub := NewHub()
	const goroutines = 50
	const roomID = int64(1)

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			client := &Client{send: make(chan []byte, 1)}

			hub.Join(roomID, client)
			hub.Broadcast(roomID, []byte("hello"))
			hub.Leave(roomID, client)
		}(i)
	}

	wg.Wait()
}
