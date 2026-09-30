package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"csp_report_backend/internal/api"
	"csp_report_backend/internal/db"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./data/csp_reports.db"
	}

	// Ensure db parent directory exists
	dbDir := filepath.Dir(dbPath)
	if dbDir != "." && dbDir != "" {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("failed to create database directory %s: %v", dbDir, err)
		}
	}

	log.Printf("Opening SQLite database at %s...", dbPath)
	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	router := api.SetupRouter(database)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Server run context for graceful shutdown
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-sig

		// Shutdown signal with grace period of 10 seconds
		shutdownCtx, shutdownCancel := context.WithTimeout(serverCtx, 10*time.Second)
		defer shutdownCancel()

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				log.Println("Graceful shutdown timed out.. forcing exit.")
			}
		}()

		log.Println("Shutting down HTTP server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		serverStopCtx()
	}()

	log.Printf("CSP Report Collector API listening on :%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("HTTP server error: %v", err)
	}

	// Wait for server context to finish
	<-serverCtx.Done()
	log.Println("Server gracefully stopped.")
}
