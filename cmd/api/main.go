package main

import (
	"log"

	"sermon-backend/internal/app"
	"sermon-backend/internal/config"
	"sermon-backend/internal/database"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	if err := database.Seed(db); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	router := app.SetupRouter(db, cfg)

	log.Printf("server started on port %s", cfg.AppPort)
	if err := router.Run(":" + cfg.AppPort); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
