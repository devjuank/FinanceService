package db

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/juank/finance-ai/backend/internal/models"
)

// Interface for DB operations
type Database interface {
	CreateUser(user models.User) error
	GetUserByEmail(email string) (models.User, error)
	GetTransactions(userID uuid.UUID) []models.Transaction
	CreateUpload(upload models.Upload) error
	GetUploads(userID uuid.UUID) []models.Upload
	UpsertTransactions(txs []models.Transaction) error
}

// Mock DB for initial development
type MemoryDB struct {
	users        map[string]models.User
	transactions map[string]models.Transaction
	uploads      []models.Upload
	mu           sync.RWMutex
}

var (
	Instance Database
	once     sync.Once
)

func GetDB() Database {
	once.Do(func() {
		// Note: main logic should call Connect() and set Instance
		if Instance == nil {
			Instance = GetMemoryDB()
		}
	})
	return Instance
}

func GetMemoryDB() *MemoryDB {
	return &MemoryDB{
		users:        make(map[string]models.User),
		transactions: make(map[string]models.Transaction),
		uploads:      []models.Upload{},
	}
}

func (db *MemoryDB) CreateUser(user models.User) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, exists := db.users[user.Email]; exists {
		return errors.New("user already exists")
	}
	db.users[user.Email] = user
	return nil
}

func (db *MemoryDB) GetUserByEmail(email string) (models.User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	user, exists := db.users[email]
	if !exists {
		return models.User{}, errors.New("user not found")
	}
	return user, nil
}

func (db *MemoryDB) GetTransactions(userID uuid.UUID) []models.Transaction {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var result []models.Transaction
	for _, tx := range db.transactions {
		if tx.UserID == userID {
			result = append(result, tx)
		}
	}
	return result
}

func (db *MemoryDB) CreateUpload(upload models.Upload) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.uploads = append(db.uploads, upload)
	return nil
}

func (db *MemoryDB) GetUploads(userID uuid.UUID) []models.Upload {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var result []models.Upload
	for _, u := range db.uploads {
		if u.UserID == userID {
			result = append(result, u)
		}
	}
	return result
}

func (db *MemoryDB) UpsertTransactions(txs []models.Transaction) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	for _, tx := range txs {
		db.transactions[tx.ID] = tx
	}
	return nil
}
