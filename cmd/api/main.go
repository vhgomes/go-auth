package main

import (
	"auth-go/internal/config"
	"auth-go/internal/handler"
	"auth-go/internal/repository"
	"auth-go/internal/service"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg, err := config.LoadPostgresConfig()
	if err != nil {
		log.Fatalf("Failed to load Postgres Config")
		return
	}

	db, err := config.InitPostgres(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres database: %v\n", err)
	}

	rediscfg, err := config.LoadRedisConfig()
	if err != nil {
		log.Fatalf("Failed to load Redis config %v\n", err)
	}

	redisClient, err := config.InitRedis(rediscfg)

	defer redisClient.Close()

	if err != nil {
		log.Fatalf("Failed to init redis: %v\n", err)
	}

	db, err = config.InitDB(db)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	router := gin.Default()

	userRepo := repository.NewUserRepository(db, redisClient)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	api := router.Group("/api/v1")
	{
		api.POST("/register", userHandler.RegisterUser)
		api.POST("/login", userHandler.LoginUser)
		api.GET("/logout", userHandler.Logout)
	}

	log.Printf("Starting server on %s\n", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
