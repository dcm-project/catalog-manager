package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/dcm-project/catalog-manager/internal/apiserver"
	"github.com/dcm-project/catalog-manager/internal/config"
	"github.com/dcm-project/catalog-manager/internal/handlers/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/service"
	"github.com/dcm-project/catalog-manager/internal/store"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := store.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create store
	dataStore := store.NewStore(db)
	defer func() {
		if err := dataStore.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Create service layer
	svc := service.NewService(dataStore)

	// Create TCP listener
	listener, err := net.Listen("tcp", cfg.Service.BindAddress)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	srv := apiserver.New(cfg, listener, v1alpha1.NewHandler(svc))

	// Create context with signal handling
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create and run server
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
