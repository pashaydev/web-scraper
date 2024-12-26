package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"web-scraper/internal/cache"
)

func CacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	stats := cache.GetInstance().GetStats()
	err := json.NewEncoder(w).Encode(stats)
	if err != nil {
		log.Println("Error encoding metrics", err.Error())
		return
	}
}

func CacheMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	metrics := cache.GetInstance().GetMetrics()
	err := json.NewEncoder(w).Encode(metrics)
	if err != nil {
		log.Println("Error encoding metrics", err.Error())
		return
	}
}
