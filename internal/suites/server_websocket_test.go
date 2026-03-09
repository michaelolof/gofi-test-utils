package suites

import (
	"fmt"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/michaelolof/gofi"
	gofiws "github.com/michaelolof/gofi/websocket"
)

func TestWebSocketEcho(t *testing.T) {
	mux := gofi.NewRouter()

	mux.Get("/ws", gofi.RouteOptions{
		Handler: gofiws.New(func(c *gofiws.Conn) error {
			for {
				mt, msg, err := c.ReadMessage()
				if err != nil {
					return err
				}
				if err = c.WriteMessage(mt, msg); err != nil {
					return err
				}
			}
			return nil
		}),
	})

	port := 38474
	addr := fmt.Sprintf(":%d", port)

	go func() {
		_ = mux.Listen(addr)
	}()

	time.Sleep(100 * time.Millisecond) // Let server start

	wsURL := fmt.Sprintf("ws://localhost:%d/ws", port)
	c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}
	defer c.Close()

	// Test 1: Echo
	err = c.WriteMessage(websocket.TextMessage, []byte("hello"))
	if err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	_, msg, err := c.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if string(msg) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", msg)
	}

	_ = mux.Shutdown()
}
