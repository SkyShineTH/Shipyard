package main

import (
	"log"
	"net/http"
	"os"

	"github.com/SkyShineTH/shipyard/auth-service/db"
	"github.com/SkyShineTH/shipyard/auth-service/handler"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8081")
	}

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	router := gin.Default()
	if err := router.SetTrustedProxies([]string{
		"127.0.0.1",
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
	}); err != nil {
		log.Fatalf("set trusted proxies: %v", err)
	}
	authHandler := handler.NewAuthHandler(database)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
	}

	port := os.Getenv("PORT")
	log.Printf("auth-service running on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
