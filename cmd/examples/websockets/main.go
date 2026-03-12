package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/michaelolof/gofi"
	"github.com/michaelolof/gofi/websocket"
)

func main() {
	r := gofi.NewRouter()
	registry := websocket.NewSessionRegistry()

	r.Get("/ws/:room_id", websocket.DefineWebSocket(websocket.WebSocketOptions{
		Handshake: websocket.HandshakePolicy{Mode: websocket.HandshakeAuto},
		Handler: func(s *websocket.Session) error {
			roomID := s.Context().Param("room_id")

			for {
				msgType, msg, err := s.ReadMessage()
				if err != nil {
					return nil
				}

				payload := []byte(fmt.Sprintf("[%s] %s", roomID, string(msg)))
				if err := s.WriteMessage(msgType, payload); err != nil {
					return err
				}
			}
		},
		Runtime: websocket.RuntimeOptions{
			Registry:        registry,
			MaxMessageBytes: 1 << 20, // 1MB
			ReadTimeout:     30 * time.Second,
			WriteTimeout:    10 * time.Second,
			Hooks: websocket.Hooks{
				OnUpgradeSuccess: func(ctx gofi.Context) {
					log.Printf("upgrade success path=%s", ctx.Path())
				},
				OnUpgradeError: func(ctx gofi.Context, err error) {
					log.Printf("upgrade error path=%s err=%v", ctx.Path(), err)
				},
				OnSessionStart: func(ctx gofi.Context) {
					log.Printf("session start room=%s active=%d", ctx.Param("room_id"), registry.Active())
				},
				OnSessionEnd: func(ctx gofi.Context, d time.Duration) {
					log.Printf("session end room=%s duration=%s active=%d", ctx.Param("room_id"), d, registry.Active())
				},
				OnSessionError: func(ctx gofi.Context, err error) {
					log.Printf("session error room=%s err=%v", ctx.Param("room_id"), err)
				},
			},
		},
	}))

	go func() {
		if err := gofi.ListenAndServe(":8080", r); err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	drainCtx, cancelDrain := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelDrain()
	if err := registry.DrainContext(drainCtx); err != nil {
		log.Printf("drain failed: %v", err)
	}

	if err := r.Shutdown(); err != nil {
		log.Printf("shutdown failed: %v", err)
	}
}
