package main

import (
	"auth-go/internal/config"
	"auth-go/internal/handler"
	"auth-go/internal/repository"
	"auth-go/internal/service"
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

	db, err := config.New(cfg)

	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}

	if err := config.InitDB(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	defer db.Close()

	log.Print("Database connected")

	router := gin.Default()

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	api := router.Group("/api/v1")
	{
		api.POST("/register", userHandler.RegisterUser) // POST /api/v1/register
	}

	fmt.Printf("Starting server on %s\n", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
