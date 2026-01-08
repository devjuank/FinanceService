package processor

import (
	"math"
	"time"

	"github.com/juank/finance-ai/backend/internal/models"
)

// NeutralizeTransfers mirrors the Python logic to find matching internal transfers
func NeutralizeTransfers(transactions []models.Transaction) []models.Transaction {
	if len(transactions) == 0 {
		return transactions
	}

	// Group transfers by amount and date
	var debits []int
	var credits []int

	for i, tx := range transactions {
		if tx.IsTransfer {
			if tx.Amount < 0 {
				debits = append(debits, i)
			} else if tx.Amount > 0 {
				credits = append(credits, i)
			}
		}
	}

	neutralizedIDs := make(map[string]bool)

	for _, dIdx := range debits {
		txD := transactions[dIdx]
		if neutralizedIDs[txD.ID] {
			continue
		}

		targetAmount := math.Abs(txD.Amount)
		dateD, _ := time.Parse("2006-01-02", txD.Date)

		for _, cIdx := range credits {
			txC := transactions[cIdx]
			if neutralizedIDs[txC.ID] {
				continue
			}

			// Similar amount (+/- 0.5%)
			if math.Abs(txC.Amount) >= targetAmount*0.995 && math.Abs(txC.Amount) <= targetAmount*1.005 {
				dateC, _ := time.Parse("2006-01-02", txC.Date)

				// Window: 1 day before to 3 days after
				diff := dateC.Sub(dateD).Hours() / 24
				if diff >= -1 && diff <= 3 {
					neutralizedIDs[txD.ID] = true
					neutralizedIDs[txC.ID] = true

					transactions[dIdx].Neutralized = true
					transactions[dIdx].Category = stringPtr("transferencia_interna")

					transactions[cIdx].Neutralized = true
					transactions[cIdx].Category = stringPtr("transferencia_interna")

					break // Found a match for this debit
				}
			}
		}
	}

	return transactions
}

func stringPtr(s string) *string {
	return &s
}
