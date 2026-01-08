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
	rows, err := db.Conn.Query(`
		SELECT id, user_id, upload_id, date, amount, source, description, merchant, category, subcategory, currency, is_transfer, is_fee, is_tax, neutralized, processed_at 
		FROM transactions WHERE user_id = $1 ORDER BY date DESC`, userID)
	if err != nil {
		return []models.Transaction{}
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		var dateStr string
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.UploadID, &dateStr, &tx.Amount, &tx.Source, &tx.Description, &tx.Merchant, &tx.Category, &tx.Subcategory, &tx.Currency, &tx.IsTransfer, &tx.IsFee, &tx.IsTax, &tx.Neutralized, &tx.ProcessedAt)
		if err == nil {
			tx.Date = dateStr
			txs = append(txs, tx)
		}
	}
	return txs
}

func (db *PostgresDB) CreateUpload(upload models.Upload) error {
	_, err := db.Conn.Exec("INSERT INTO uploads (id, user_id, filename, status, created_at) VALUES ($1, $2, $3, $4, $5)",
		upload.ID, upload.UserID, upload.Filename, upload.Status, upload.CreatedAt)
	return err
}

func (db *PostgresDB) GetUploads(userID uuid.UUID) []models.Upload {
	rows, err := db.Conn.Query("SELECT id, user_id, filename, status, created_at FROM uploads WHERE user_id = $1 ORDER BY created_at DESC", userID)
	if err != nil {
		return []models.Upload{}
	}
	defer rows.Close()

	var uploads []models.Upload
	for rows.Next() {
		var u models.Upload
		if err := rows.Scan(&u.ID, &u.UserID, &u.Filename, &u.Status, &u.CreatedAt); err == nil {
			uploads = append(uploads, u)
		}
	}
	return uploads
}

func (db *PostgresDB) UpsertTransactions(txs []models.Transaction) error {
	for _, tx := range txs {
		_, err := db.Conn.Exec(`
			INSERT INTO transactions (id, user_id, upload_id, date, amount, source, description, merchant, category, subcategory, currency, is_transfer, is_fee, is_tax, neutralized, processed_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
			ON CONFLICT (id) DO UPDATE SET
				upload_id = EXCLUDED.upload_id,
				category = EXCLUDED.category,
				subcategory = EXCLUDED.subcategory,
				merchant = EXCLUDED.merchant,
				is_transfer = EXCLUDED.is_transfer,
				is_fee = EXCLUDED.is_fee,
				is_tax = EXCLUDED.is_tax,
				neutralized = EXCLUDED.neutralized
		`, tx.ID, tx.UserID, tx.UploadID, tx.Date, tx.Amount, tx.Source, tx.Description, tx.Merchant, tx.Category, tx.Subcategory, tx.Currency, tx.IsTransfer, tx.IsFee, tx.IsTax, tx.Neutralized, tx.ProcessedAt)
		if err != nil {
			return err
		}
	}
	return nil
}
