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

	// Public routes
	mux.HandleFunc("/api/login", middleware.ChainMiddleware(
		handlers.Login,
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
	))
	mux.HandleFunc("/api/signup", middleware.ChainMiddleware(
		handlers.Signup,
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
	))
	mux.HandleFunc("/health", handlers.HealthCheckHandler)

	// Protected routes
	mux.HandleFunc("/api/scraper", middleware.ChainMiddleware(
		handlers.SearchHandler,
		middleware.AuthMiddleware,
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
	))
	mux.HandleFunc("/api/scraper-deep", middleware.ChainMiddleware(
		handlers.SearchDeepHandler,
		middleware.AuthMiddleware,
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
	))
	mux.HandleFunc("/cache/stats", middleware.ChainMiddleware(
		handlers.CacheStatsHandler,
		middleware.AuthMiddleware,
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
	))
	mux.HandleFunc("/cache/metrics", middleware.ChainMiddleware(
		handlers.CacheMetricsHandler,
		middleware.AuthMiddleware,
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
	))

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
