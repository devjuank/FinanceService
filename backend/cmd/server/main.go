package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"encoding/json"
	"io"
	"path/filepath"
	"time"

	"strings"

	"github.com/google/uuid"
	"github.com/juank/finance-ai/backend/internal/api"
	"github.com/juank/finance-ai/backend/internal/auth"
	"github.com/juank/finance-ai/backend/internal/db"
	"github.com/juank/finance-ai/backend/internal/models"
	"github.com/juank/finance-ai/backend/internal/processor"
	"github.com/juank/finance-ai/backend/internal/processor/common"
	"github.com/juank/finance-ai/backend/internal/processor/parsers"
)

func main() {
	// Initialize Database
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	db.Instance = database

	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("/api/login", handleLogin)
	mux.HandleFunc("/api/register", handleRegister)

	// Protected routes
	mux.HandleFunc("/api/upload", api.AuthMiddleware(handleUpload))
	mux.HandleFunc("/api/transactions", api.AuthMiddleware(handleTransactions))

	fmt.Println("Server starting on :8080...")
	log.Fatal(http.ListenAndServe(":8080", api.CORSMiddleware(mux)))
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

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	// Max 10MB
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to get file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create Upload Record
	uploadID := uuid.New()
	upload := models.Upload{
		ID:        uploadID,
		UserID:    userID,
		Filename:  handler.Filename,
		Status:    "processing",
		CreatedAt: time.Now(),
	}
	db.GetDB().CreateUpload(upload)

	// Save file temporarily
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%s", uploadID, handler.Filename))
	dst, err := os.Create(tempPath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tempPath)
	io.Copy(dst, file)
	dst.Close()

	// Pick Parser
	parser := pickParser(handler.Filename)
	if parser == nil {
		http.Error(w, "Unsupported file type or bank", http.StatusBadRequest)
		return
	}

	// Trigger processing
	outputDir := "/Users/juank/Documents/Cuentas/DatosClasificados"
	engine := processor.NewEngine(outputDir, userID)
	txs, err := engine.ProcessFile(tempPath, parser, uploadID)
	if err != nil {
		http.Error(w, "Processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := engine.SaveAndConsolidate(txs); err != nil {
		http.Error(w, "Save failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	api.JSONResponse(w, http.StatusOK, map[string]interface{}{
		"message":   "File processed successfully",
		"upload_id": uploadID,
		"count":     len(txs),
	})
}

func pickParser(filename string) common.Normalizer {
	ext := strings.ToLower(filepath.Ext(filename))
	name := strings.ToLower(filename)

	if ext == ".pdf" {
		if strings.Contains(name, "brubank") {
			return &parsers.BrubankPDFParser{}
		}
		if strings.Contains(name, "visa") || strings.Contains(name, "santander") {
			return &parsers.SantanderVisaPDFParser{}
		}
	} else if ext == ".csv" {
		if strings.Contains(name, "mercadopago") {
			return &parsers.MercadoPagoParser{}
		}
		if strings.Contains(name, "deel") {
			return &parsers.DeelParser{}
		}
	} else if ext == ".xlsx" {
		if strings.Contains(name, "santander") {
			return &parsers.SantanderXLSXParser{}
		}
	}
	return nil
}

func handleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.Header.Get("X-User-ID")
	userID, _ := uuid.Parse(userIDStr)

	txs := db.GetDB().GetTransactions(userID)
	api.JSONResponse(w, http.StatusOK, txs)
}
