package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// LoggingMiddleware logs all requests
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next(w, r)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, time.Since(start))
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			respondWithError(w, "No authorization header", http.StatusUnauthorized)
			return
		}

		// Check if the header starts with "Bearer "
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			respondWithError(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := bearerToken[1]
		claims := jwt.MapClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			respondWithError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			respondWithError(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		r = r.WithContext(setUserClaimsToContext(r.Context(), claims))
		next(w, r)
	}
}

func respondWithError(w http.ResponseWriter, message string, status int) {
	response := AuthResponse{
		Success: false,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// Helper function to combine multiple middleware
func ChainMiddleware(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
