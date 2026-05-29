package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	_ = godotenv.Load()

	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", "8082")
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
	router.Use(metricsMiddleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	client, err := NewInClusterClient()
	if err != nil {
		log.Printf("platform snapshot client unavailable: %v", err)
	}

	var provider SnapshotProvider = unavailableProvider{}
	if client != nil {
		provider = NewKubeSnapshotProvider(client, SnapshotConfigFromEnv())
	}

	api := router.Group("/api/v1")
	{
		api.GET("/platform/status", PlatformStatusHandler(provider))
	}

	port := os.Getenv("PORT")
	log.Printf("platform-status-service running on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
