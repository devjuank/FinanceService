package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/juank/finance-ai/backend/internal/models"
	_ "github.com/lib/pq"
)

type PostgresDB struct {
	Conn *sql.DB
}

func Connect() (Database, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		// Fallback to memory for local dev without docker
		return GetMemoryDB(), nil
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresDB{Conn: db}, nil
}

func (db *PostgresDB) CreateUser(user models.User) error {
	_, err := db.Conn.Exec("INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)",
		user.ID, user.Email, user.PasswordHash)
	return err
}

func (db *PostgresDB) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	err := db.Conn.QueryRow("SELECT id, email, password_hash FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (db *PostgresDB) GetTransactions(userID uuid.UUID) []models.Transaction {
	// For now, return empty or implement query
	return []models.Transaction{}
}
