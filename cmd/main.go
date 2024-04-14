package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/BelyaevEI/backend-trainee-assignment-2024/internal/app"
)

func main() {

	// Init application
	app, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	// Start server for proccesiing request
	go func() {
		log.Println("Server is start")
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe error: %v", err)
		}
	}()

	// Given signal for shutdown
	sig := <-app.Sigint
	log.Printf("Received signal: %v", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := app.Server.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

}
