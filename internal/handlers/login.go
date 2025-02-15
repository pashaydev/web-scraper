package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"net/http"
	"os"
	"time"
	"web-scraper/internal/database"
	"web-scraper/internal/models"
)

type SignupRequest struct {
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	RepeatPassword string `json:"repeatPassword" binding:"required"`
	Email          string `json:"email" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token   string `json:"token,omitempty"`
	Success bool   `json:"success,omitempty"`
	Message string `json:"message,omitempty"`
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Failed to marshal JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// Helper functions remain the same
func generateUserId() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%d", rand.Intn(1000000))
}

func generateToken(user models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.UserID,
		"username": user.Username,
		"email":    user.Email,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Password != req.RepeatPassword {
		respondWithJSON(w, http.StatusUnprocessableEntity, AuthResponse{
			Success: false,
			Message: "Passwords do not match",
		})
		return
	}

	ip := r.Header.Get("X-Forwarded-For")
	db := database.GetDB()

	// Check if username exists
	data, _, err := db.Client.From("users").Select("*", "", false).Eq("username", req.Username).Execute()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	var existingUsers []models.User
	if err := json.Unmarshal(data, &existingUsers); err != nil {
		http.Error(w, "Failed to parse database response", http.StatusInternalServerError)
		return
	}

	if len(existingUsers) > 0 {
		respondWithJSON(w, http.StatusUnprocessableEntity, AuthResponse{
			Success: false,
			Message: "User with this username already exists",
		})
		return
	}

	// Check if IP exists
	//data, _, err = db.Client.From("users").Select("*", "", false).Eq("ip", ip).Execute()
	//if err != nil {
	//	http.Error(w, "Database error", http.StatusInternalServerError)
	//	return
	//}
	//
	//var existingIPUsers []models.User
	//if err := json.Unmarshal(data, &existingIPUsers); err != nil {
	//	http.Error(w, "Failed to parse database response", http.StatusInternalServerError)
	//	return
	//}
	//
	//if len(existingIPUsers) > 0 {
	//	respondWithJSON(w, http.StatusUnauthorized, AuthResponse{
	//		Success: false,
	//		Message: "IP should be unique",
	//	})
	//	return
	//}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	userId := generateUserId()

	// Create new user with only the required fields
	newUser := map[string]interface{}{
		"username":  req.Username,
		"user_id":   userId,
		"hash_pass": string(hashedPassword),
		"from":      "web",
		"chat_id":   userId,
		"email":     req.Email,
		"ip":        ip,
		// Don't include created_at as it's handled by the database
		// Don't include id as it's auto-generated
		"active_command": nil,
	}

	// Insert new user
	_, _, err = db.Client.From("users").Insert(newUser, true, "", "*", "").Execute()
	if err != nil {
		fmt.Printf("Insert error: %v\n", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Fetch the newly created user
	data, _, err = db.Client.From("users").Select("*", "", false).Eq("user_id", userId).Execute()
	if err != nil {
		http.Error(w, "Failed to fetch created user", http.StatusInternalServerError)
		return
	}
	var insertedUsers []models.User
	if err := json.Unmarshal(data, &insertedUsers); err != nil {
		http.Error(w, "Failed to parse database response", http.StatusInternalServerError)
		return
	}

	if len(insertedUsers) == 0 {
		http.Error(w, "Failed to fetch created user", http.StatusInternalServerError)
		return
	}

	token, err := generateToken(insertedUsers[0])
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Token:   token,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := database.GetDB()

	data, _, err := db.Client.From("users").Select("*", "", false).Eq("username", req.Username).Execute()
	if err != nil {
		respondWithJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		http.Error(w, "Failed to parse database response", http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		respondWithJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	user := users[0]
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashPass), []byte(req.Password)); err != nil {
		respondWithJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid credentials",
		})
		return
	}

	token, err := generateToken(user)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Token:   token,
	})
}
