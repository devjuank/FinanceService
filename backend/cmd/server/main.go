package main

import (
	"fmt"
	"log"
	"net/http"

	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/juank/finance-ai/backend/internal/api"
	"github.com/juank/finance-ai/backend/internal/auth"
	"github.com/juank/finance-ai/backend/internal/db"
	"github.com/juank/finance-ai/backend/internal/models"
)

func main() {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/register", handleRegister)

	// Protected routes
	mux.HandleFunc("/api/upload", api.AuthMiddleware(handleUpload))
	mux.HandleFunc("/api/transactions", api.AuthMiddleware(handleTransactions))

	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	user, err := db.GetDB().GetUserByEmail(req.Email)
	if err != nil || !auth.CheckPasswordHash(req.Password, user.PasswordHash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	api.JSONResponse(w, http.StatusOK, map[string]string{"token": token})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hashed, _ := auth.HashPassword(req.Password)
	user := models.User{
		ID:           uuid.New(),
		Email:        req.Email,
		PasswordHash: hashed,
		CreatedAt:    time.Now(),
	}

	if err := db.GetDB().CreateUser(user); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	api.JSONResponse(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID := r.Header.Get("X-User-ID")
	fmt.Printf("Uploading file for user: %s\n", userID)
	api.JSONResponse(w, http.StatusOK, map[string]string{"message": "File upload started (mock)"})
}

func handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	api.JSONResponse(w, http.StatusOK, []string{})
}
