package main

import (
	"auth-go/internal/config"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	dbAddr := os.Getenv("DB_ADDR")

	cfg, err := config.LoadConfig(port, dbAddr)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	router := gin.Default()

	db, err := config.New(cfg)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}

	defer db.Close()

	log.Print("Database connected")

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the application!",
		})
	})

	fmt.Printf("Starting server on %s\n", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
