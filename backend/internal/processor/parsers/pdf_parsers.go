package parsers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/juank/finance-ai/backend/internal/models"
	"github.com/juank/finance-ai/backend/internal/processor/common"
	"github.com/ledongthuc/pdf"
)

type BrubankPDFParser struct{}

func (p *BrubankPDFParser) Normalize(filePath string) ([]models.Transaction, error) {
	content, err := readPDFText(filePath)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(content, "\n")
	var transactions []models.Transaction
	dateRegex := regexp.MustCompile(`^\d{2}-\d{2}-\d{2}$`)

	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if dateRegex.MatchString(line) {
			if i+5 >= len(lines) {
				continue
			}

			dateRaw := line
			t, err := time.Parse("02-01-06", dateRaw)
			if err != nil {
				continue
			}
			dateISO := t.Format("2006-01-02")

			// Brubank format:
			// idx+1: Reference (ignored)
			// idx+2: Description
			// idx+3: Debit
			// idx+4: Credit
			// idx+5: Balance
			description := strings.TrimSpace(lines[i+2])
			debit := common.CleanAmount(lines[i+3])
			credit := common.CleanAmount(lines[i+4])
			balance := common.CleanAmount(lines[i+5])

			amount := credit - debit
			direction := "debit"
			if amount > 0 {
				direction = "credit"
			}

			isTax := containsAny(description, "iva", "percepción", "ganancias", "impuesto", "sircreb", "arca", "afip")
			isTransfer := containsAny(description, "cuenta tuya", "transferencia", "enviada", "recibida")
			isFee := containsAny(description, "comisión", "reimpresión", "intereses pagados", "mantenimiento")

			cat, sub := common.InferCategory(description)

			merchant := ""
			if !isTax && !isTransfer && !isFee {
				parts := strings.Split(description, " ")
				if len(parts) > 0 {
					merchant = parts[0]
				}
			}

			var merchantPtr *string
			if merchant != "" {
				merchantPtr = &merchant
			}

			transactions = append(transactions, models.Transaction{
				ID:          common.GenerateID("brubank", "caja_ahorro_pesos", dateISO, fmt.Sprintf("%.2f", amount), description),
				Source:      "brubank",
				Account:     "caja_ahorro_pesos",
				Date:        dateISO,
				Amount:      amount,
				Currency:    "ARS",
				Description: description,
				Direction:   direction,
				Merchant:    merchantPtr,
				Category:    cat,
				Subcategory: sub,
				Balance:     &balance,
				IsTransfer:  isTransfer,
				IsFee:       isFee,
				IsTax:       isTax,
			})

			i += 5
		}
	}

	return transactions, nil
}

type SantanderVisaPDFParser struct{}

func (p *SantanderVisaPDFParser) Normalize(filePath string) ([]models.Transaction, error) {
	content, err := readPDFText(filePath)
	if err != nil {
		return nil, err
	}

	monthsMap := map[string]string{
		"Enero": "01", "Febrero": "02", "Marzo": "03", "Abril": "04", "Mayo": "05", "Junio": "06",
		"Julio": "07", "Agosto": "08", "Setiembre": "09", "Septiembre": "09", "Octubre": "10", "Noviembre": "11", "Diciembre": "12",
		"Ene": "01", "Feb": "02", "Mar": "03", "Abr": "04", "May": "05", "Jun": "06",
		"Jul": "07", "Ago": "08", "Set": "09", "Sep": "09", "Oct": "10", "Nov": "11", "Dic": "12",
	}

	// Find year in header
	year := "2025"
	yearRegex := regexp.MustCompile(`CIERRE\s+\d{2}\s+\w{3}\s+(\d{2})`)
	if match := yearRegex.FindStringSubmatch(content); len(match) > 1 {
		year = "20" + match[1]
	}

	txRegex := regexp.MustCompile(`(\d{2})\s+([a-zA-Z]{3,10})\s+(\d{2})\s+.*?\s+(.*?)\s+([\d\.\,]+-?)(\s+[\d\.\,]+-?)?$`)
	lines := strings.Split(content, "\n")
	var transactions []models.Transaction

	for _, line := range lines {
		match := txRegex.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		monthStr := strings.Title(strings.ToLower(match[2]))
		dayTx := match[3]
		description := strings.TrimSpace(match[4])
		amountStr := match[5]

		month, ok := monthsMap[monthStr]
		if !ok {
			continue
		}

		dateISO := fmt.Sprintf("%s-%s-%s", year, month, dayTx)

		isNegative := false
		if strings.HasSuffix(amountStr, "-") {
			isNegative = true
			amountStr = strings.TrimSuffix(amountStr, "-")
		}

		amount := common.CleanAmount(amountStr)
		direction := "debit"
		if isNegative {
			amount = float64(int(amount*100)) / 100 // positive on CC usually means credit/payment
			direction = "credit"
		} else {
			amount = -amount
			direction = "debit"
		}

		if amount == 0 {
			continue
		}

		isTax := containsAny(description, "impuesto", "iva", "percepción", "db.rg")
		descUpper := strings.ToUpper(description)
		isTransfer := strings.Contains(descUpper, "SU PAGO") || strings.Contains(descUpper, "PAGO EN")
		isFee := containsAny(description, "comision", "cargo", "interes")

		cat, sub := common.InferCategory(description)

		merchant := strings.Split(description, " ")[0]
		var merchantPtr *string
		if merchant != "" {
			merchantPtr = &merchant
		}

		transactions = append(transactions, models.Transaction{
			ID:          common.GenerateID("santander", "credito_visa", dateISO, fmt.Sprintf("%.2f", amount), description),
			Source:      "santander",
			Account:     "credito_visa",
			Date:        dateISO,
			Amount:      amount,
			Currency:    "ARS",
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

func readPDFText(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buf bytes.Buffer
	for i := 1; i <= r.NumPage(); i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		text, _ := p.GetPlainText(nil)
		buf.WriteString(text)
		buf.WriteString("\n")
	}
	return buf.String(), nil
}
