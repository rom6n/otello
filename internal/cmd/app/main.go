package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/rom6n/otello/docs"
	"github.com/rom6n/otello/internal/app/config"
	"github.com/rom6n/otello/internal/pkg/database"
	"github.com/rom6n/otello/internal/pkg/http"
)

func main() {
	ctx := context.Background()
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}

	dbClient := database.NewClient()
	configs := config.GetConfig(dbClient)

	app := http.NewFiberApp(configs)

	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080" // default
		}
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("error starting app: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	<-stop
	shutdownTimeSecond := 3 * time.Second // mainly 35s
	shutdownTime := 4                     // mainly 40s

	ctxShutdown, cancel := context.WithTimeout(ctx, shutdownTimeSecond)
	defer cancel()

	if shotdownErr := app.ShutdownWithContext(ctxShutdown); shotdownErr != nil {
		log.Fatalf("error shutting down server: %v. forced shutdown", shotdownErr)
	}

	for i := shutdownTime; i > 0; i -= 1 {
		log.Printf("shutting down in %v seconds...\n", i)
		time.Sleep(1 * time.Second)
	}

	log.Println("server shutdown successfully")
}
