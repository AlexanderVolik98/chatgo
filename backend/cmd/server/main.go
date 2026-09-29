package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"chatgo/backend/internal/auth"
	"chatgo/backend/internal/config"
	"chatgo/backend/internal/db"
	"chatgo/backend/internal/httpapi"
	"chatgo/backend/internal/repository"
	"chatgo/backend/internal/ws"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	tokens := auth.NewTokenManager(cfg.JWTSecret)
	users := repository.NewUserRepository(pool)
	rooms := repository.NewRoomRepository(pool)
	msgs := repository.NewMessageRepository(pool)

	hub := ws.NewHub()
	wsService := ws.NewService(hub, rooms, msgs, tokens)

	router := httpapi.NewRouter(httpapi.Deps{
		Tokens: tokens,
		Users:  users,
		Rooms:  rooms,
		Msgs:   msgs,
		WS:     wsService,
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("chatgo server listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
