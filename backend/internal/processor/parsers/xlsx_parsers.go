package parsers

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/juank/finance-ai/backend/internal/models"
	"github.com/juank/finance-ai/backend/internal/processor/common"
	"github.com/xuri/excelize/v2"
)

type SantanderXLSXParser struct{}

func (p *SantanderXLSXParser) Normalize(filePath string) ([]models.Transaction, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Get all the rows in the Sheet1
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		// Try first sheet if Sheet1 fails
		sheets := f.GetSheetList()
		if len(sheets) > 0 {
			rows, err = f.GetRows(sheets[0])
		}
		if err != nil {
			return nil, err
		}
	}

	var transactions []models.Transaction
	dateRegex := regexp.MustCompile(`\d{2}/\d{2}/\d{4}`)

	for i := 12; i < len(rows); i++ {
		row := rows[i]
		if len(row) < 8 {
			continue
		}

		dateVal := row[1]
		if !dateRegex.MatchString(dateVal) {
			continue
		}

		// Format: DD/MM/YYYY to YYYY-MM-DD
		parts := strings.Split(dateVal, "/")
		if len(parts) != 3 {
			continue
		}
		dateISO := fmt.Sprintf("%s-%s-%s", parts[2], parts[1], parts[0])

		description := strings.TrimSpace(row[3])
		amount := common.CleanAmount(row[6])
		balance := common.CleanAmount(row[7])

		if amount == 0 {
			continue
		}

		direction := "debit"
		if amount > 0 {
			direction = "credit"
		}

		isTax := containsAny(description, "impuesto", "iva", "percepción", "sircreb", "db.rg")
		isTransfer := strings.Contains(strings.ToLower(description), "transferencia")
		isFee := containsAny(description, "comision", "cargo", "interes")

		cat, sub := common.InferCategory(description)

		transactions = append(transactions, models.Transaction{
			ID:          common.GenerateID("santander", "caja_ahorro_pesos", dateISO, fmt.Sprintf("%.2f", amount), description),
			Source:      "santander",
			Account:     "caja_ahorro_pesos",
			Date:        dateISO,
			Amount:      amount,
			Currency:    "ARS",
			Description: description,
			Direction:   direction,
			Category:    cat,
			Subcategory: sub,
			Balance:     &balance,
			IsTransfer:  isTransfer,
			IsFee:       isFee,
			IsTax:       isTax,
		})
	}

	return transactions, nil
}
