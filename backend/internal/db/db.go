package db

import (
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/juank/finance-ai/backend/internal/models"
)

// Mock DB for initial development
type MemoryDB struct {
	users        map[string]models.User
	transactions []models.Transaction
	mu           sync.RWMutex
}

var (
	Instance *MemoryDB
	once     sync.Once
)

func GetDB() *MemoryDB {
	once.Do(func() {
		Instance = &MemoryDB{
			users:        make(map[string]models.User),
			transactions: []models.Transaction{},
		}
	})
	return Instance
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
