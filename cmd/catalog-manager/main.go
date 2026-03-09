package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/dcm-project/catalog-manager/internal/apiserver"
	"github.com/dcm-project/catalog-manager/internal/config"
	"github.com/dcm-project/catalog-manager/internal/handlers/v1alpha1"
	"github.com/dcm-project/catalog-manager/internal/placement"
	"github.com/dcm-project/catalog-manager/internal/service"
	"github.com/dcm-project/catalog-manager/internal/store"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		return 1
	}

	// Initialize database
	db, err := store.InitDB(cfg)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		return 1
	}

	// Create store
	dataStore := store.NewStore(db)
	defer func() {
		if err := dataStore.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Create Placement Manager client
	pmClient, err := placement.NewClient(cfg.Placement.URL)
	if err != nil {
		log.Printf("Failed to create placement manager client: %v", err)
		return 1
	}

	// Create context with signal handling (used for seed and server)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Create service layer
	svc := service.NewService(dataStore, pmClient)

	// Seed service types and default catalog items if empty
	if err := svc.Seed(ctx); err != nil {
		log.Printf("Failed to seed database: %v", err)
		return 1
	}

	// Create TCP listener
	listener, err := net.Listen("tcp", cfg.Service.BindAddress)
	if err != nil {
		log.Printf("Failed to create listener: %v", err)
		return 1
	}
	defer func() { _ = listener.Close() }()

	srv := apiserver.New(cfg, listener, v1alpha1.NewHandler(svc))

	// Run server
	if err := srv.Run(ctx); err != nil {
		log.Printf("Server failed: %v", err)
		return 1
	}

	return 0
}
