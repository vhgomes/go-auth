package main

import (
	"auth-go/internal/config"
	"auth-go/internal/handler"
	"auth-go/internal/repository"
	"auth-go/internal/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {

	cfg, err := config.LoadPostgresConfig()
	if err != nil {
		fmt.Printf("Failed to load Postgress config: %v\n", err)
		return
	}

	db, err := config.InitPostgres(cfg)
	if err != nil {
		fmt.Printf("Failed to connect to Postgres database: %v\n", err)
	}

	rediscfg, err := config.LoadRedisConfig()
	if err != nil {
		fmt.Printf("Failed to load Redis config %v\n", err)
	}

	redisClient, err := config.InitRedis(rediscfg)
	defer redisClient.Close()

	if err != nil {
		fmt.Printf("Failed to init redis: %v\n", err)
	}

	if err := config.InitDB(db); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Print("Database connected")

	router := gin.Default()

	userRepo := repository.NewUserRepository(db, redisClient)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	api := router.Group("/api/v1")
	{
		api.POST("/register", userHandler.RegisterUser) // POST /api/v1/register
		api.POST("/login", userHandler.LoginUser)
	}

	fmt.Printf("Starting server on %s\n", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
