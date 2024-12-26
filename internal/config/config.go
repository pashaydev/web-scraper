package config

import "time"

type Configuration struct {
	Port          string
	RateLimit     int
	CacheDuration time.Duration
	MaxCacheSize  int
	MaxCacheBytes int64
}

var Config = Configuration{
	Port:          "8080",
	RateLimit:     5,
	CacheDuration: 5 * time.Minute,
	MaxCacheSize:  1000,
	MaxCacheBytes: 50 * 1024 * 1024,
}
