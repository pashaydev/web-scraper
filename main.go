package main

import (
	"log"
	"net/http"
	"time"
	"web-scraper/internal/config"
	"web-scraper/internal/handlers"
	"web-scraper/internal/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/login", middleware.RecoveryMiddleware(middleware.LoggingMiddleware(handlers.Login)))
	mux.HandleFunc("/api/signup", middleware.RecoveryMiddleware(middleware.LoggingMiddleware(handlers.Signup)))
	mux.HandleFunc("/api/scraper", middleware.RecoveryMiddleware(middleware.LoggingMiddleware(handlers.SearchHandler)))
	mux.HandleFunc("/api/scraper-deep", middleware.RecoveryMiddleware(middleware.LoggingMiddleware(handlers.SearchDeepHandler)))
	mux.HandleFunc("/health", handlers.HealthCheckHandler)
	mux.HandleFunc("/cache/stats", handlers.CacheStatsHandler)
	mux.HandleFunc("/cache/metrics", handlers.CacheMetricsHandler)

	server := &http.Server{
		Addr:         ":" + config.Config.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server starting on :%s...", config.Config.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
