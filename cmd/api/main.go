package main

import (
	"auth-go/internal/config"
	"auth-go/internal/handler"
	"auth-go/internal/middleware"
	"auth-go/internal/repository"
	"auth-go/internal/service"
	"log"
	"time"

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

	userRepo := repository.NewPgUserRepository(db, redisClient)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService, cfg)
	rateLimiter := middleware.NewRateLimiter(redisClient, 5, time.Minute) // 5 tentativas/minuto por IP

	api := router.Group("/api/v1")
	{
		public := api.Group("/")
		public.Use(rateLimiter.Limit())
		{
			public.POST("/register", userHandler.RegisterUser)
			public.POST("/login", userHandler.LoginUser)
		}

		protected := api.Group("/")
		protected.Use(rateLimiter.Limit(), middleware.RequireAuth(userRepo))
		{
			protected.GET("/logout", userHandler.Logout)
		}
	}
	log.Printf("Starting server on %s\n", cfg.Addr)
	if err := router.Run(cfg.Addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
