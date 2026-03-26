package main

import (
	"log"
	"net/http"
	"os"

	"github.com/SkyShineTH/shipyard/todo-service/db"
	"github.com/SkyShineTH/shipyard/todo-service/handler"
	"github.com/SkyShineTH/shipyard/todo-service/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8080")
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
	todoHandler := handler.NewTodoHandler(database)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := router.Group("/api/v1")
	api.Use(middleware.RequireAuth())
	{
		api.GET("/todos", todoHandler.GetTodos)
		api.POST("/todos", todoHandler.CreateTodo)
		api.PUT("/todos/:id", todoHandler.UpdateTodo)
		api.DELETE("/todos/:id", todoHandler.DeleteTodo)
	}

	port := os.Getenv("PORT")
	log.Printf("todo-service running on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
