package parsers

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/juank/finance-ai/backend/internal/models"
	"github.com/juank/finance-ai/backend/internal/processor/common"
)

type MercadoPagoParser struct{}

func (p *MercadoPagoParser) Normalize(filePath string) ([]models.Transaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Handle different separators and headers
	// MP CSV usually has a header with RELEASE_DATE
	// We'll read line by line to find the header

	rawLines, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(rawLines), "\n")
	headerIdx := -1
	sep := ","

	for i, line := range lines {
		if strings.Contains(line, "RELEASE_DATE") {
			headerIdx = i
			if strings.Contains(line, ";") {
				sep = ";"
			}
			break
		}
	}

	if headerIdx == -1 {
		return nil, fmt.Errorf("could not find MercadoPago header")
	}

	// Create a new reader from the header onwards
	reader := csv.NewReader(strings.NewReader(strings.Join(lines[headerIdx:], "\n")))
	reader.Comma = rune(sep[0])
	reader.LazyQuotes = true

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, nil
	}

	headers := records[0]
	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[h] = i
	}

	var transactions []models.Transaction
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < len(headers) {
			continue
		}

		dateVal := row[colMap["RELEASE_DATE"]]
		if !regexp.MustCompile(`\d{2}-\d{2}-\d{4}`).MatchString(dateVal) {
			continue
		}

		t, _ := time.Parse("02-01-2006", dateVal)
		dateISO := t.Format("2006-01-02")

		amount := common.CleanAmount(row[colMap["TRANSACTION_NET_AMOUNT"]])
		description := row[colMap["TRANSACTION_TYPE"]]
		balance := common.CleanAmount(row[colMap["PARTIAL_BALANCE"]])

		direction := "debit"
		if amount > 0 {
			direction = "credit"
		}

		isTax := containsAny(description, "percepción", "iva", "impuesto", "arca")
		isTransfer := containsAny(description, "transferencia", "enviaste", "recibiste", "de una cuenta tuya", "a una cuenta tuya")
		isFee := strings.Contains(strings.ToLower(description), "comisión")

		cat, sub := common.InferCategory(description)

		merchant := ""
		if strings.HasPrefix(description, "Pago ") {
			merchant = strings.Split(strings.TrimPrefix(description, "Pago "), " ")[0]
		} else if strings.HasPrefix(description, "Transferencia ") {
			merchant = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(description, "Transferencia ", ""), "enviada ", ""), "recibida ", "")
		}

		var merchantPtr, catPtr, subPtr *string
		if merchant != "" {
			merchantPtr = &merchant
		}
		catPtr = cat
		subPtr = sub

		transactions = append(transactions, models.Transaction{
			ID:          common.GenerateID("mercadopago", "cuenta_digital", dateISO, fmt.Sprintf("%.2f", amount), description),
			Source:      "mercadopago",
			Account:     "cuenta_digital",
			Date:        dateISO,
			Amount:      amount,
			Currency:    "ARS",
			Description: description,
			Direction:   direction,
			Merchant:    merchantPtr,
			Category:    catPtr,
			Subcategory: subPtr,
			Balance:     &balance,
			IsTransfer:  isTransfer,
			IsFee:       isFee,
			IsTax:       isTax,
		})
	}

	return transactions, nil
}

type DeelParser struct{}

func (p *DeelParser) Normalize(filePath string) ([]models.Transaction, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, nil
	}

	headers := records[0]
	colMap := make(map[string]int)
	for i, h := range headers {
		colMap[h] = i
	}

	var transactions []models.Transaction
	for i := 1; i < len(records); i++ {
		row := records[i]

		status := strings.ToLower(row[colMap["Transaction Status"]])
		if status != "completed" {
			continue
		}

		dateVal := row[colMap["Date Requested"]]
		dateISO := strings.Split(dateVal, " ")[0]

		amountStr := row[colMap["Transaction Amount"]]
		amount, _ := strconv.ParseFloat(amountStr, 64)
		currency := row[colMap["Currency"]]
		txType := row[colMap["Transaction Type"]]

		client := row[colMap["Client"]]
		contract := row[colMap["Contract Name"]]
		description := strings.TrimSpace(fmt.Sprintf("%s: %s %s", txType, client, contract))

		direction := "debit"
		if amount > 0 {
			direction = "credit"
		}

		isTax := strings.Contains(strings.ToLower(description), "tax")
		isTransfer := txType == "withdrawal" || txType == "deel_card_withdrawal"
		isFee := strings.Contains(strings.ToLower(description), "fee") || txType == "provider_fee"

		cat, sub := common.InferCategory(description)
		if cat == nil && txType == "client_payment" {
			c, s := "ingresos", "sueldo"
			cat, sub = &c, &s
		}

		var merchantPtr *string
		if client != "" {
			merchantPtr = &client
		}

		transactions = append(transactions, models.Transaction{
			ID:          common.GenerateID("deel", "balance_usd", dateISO, fmt.Sprintf("%.2f", amount), description),
			Source:      "deel",
			Account:     "balance_usd",
			Date:        dateISO,
			Amount:      amount,
			Currency:    currency,
			Description: description,
			Direction:   direction,
			Merchant:    merchantPtr,
			Category:    cat,
			Subcategory: sub,
			IsTransfer:  isTransfer,
			IsFee:       isFee,
			IsTax:       isTax,
		})
	}

	return transactions, nil
}

func containsAny(s string, keywords ...string) bool {
	lower := strings.ToLower(s)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

func strconvParseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}
