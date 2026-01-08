package common

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/juank/finance-ai/backend/internal/models"
)

// Normalizer defines the interface for different bank report parsers
type Normalizer interface {
	Normalize(filePath string) ([]models.Transaction, error)
}

// GenerateID creates a deterministic transaction_id
func GenerateID(source, account, date, amount, rawDescription string) string {
	payload := fmt.Sprintf("%s%s%s%s%s", source, account, date, amount, rawDescription)
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}

// CleanAmount parses a string into a float64, handling various formatting styles
func CleanAmount(amountStr string) float64 {
	if amountStr == "" || amountStr == "-" || strings.ToLower(amountStr) == "nan" {
		return 0.0
	}

	// Remove currency symbols and other non-numeric chars except . , -
	re := regexp.MustCompile(`[^\d,\.\-]`)
	clean := re.ReplaceAllString(amountStr, "")
	clean = strings.TrimSpace(clean)
	if clean == "" {
		return 0.0
	}

	// Handle ARG/International formats: 1.234,56 or 1234,56
	if strings.Contains(clean, ",") && strings.Contains(clean, ".") {
		clean = strings.ReplaceAll(clean, ".", "")
		clean = strings.ReplaceAll(clean, ",", ".")
	} else if strings.Contains(clean, ",") {
		clean = strings.ReplaceAll(clean, ",", ".")
	}

	val, err := strconv.ParseFloat(clean, 64)
	if err != nil {
		return 0.0
	}
	return val
}

// InferCategory determines category and subcategory based on description
func InferCategory(description string) (category *string, subcategory *string) {
	desc := strings.ToLower(description)

	rules := []struct {
		keywords []string
		cat      string
		sub      string
	}{
		{[]string{"iva", "percepción", "ganancias", "tax", "impuesto", "sircreb", "arca", "afip"}, "impuestos", "impuestos y contribuciones"},
		{[]string{"netflix", "spotify", "youtube", "primevideo", "disney", "steam"}, "entretenimiento", "servicios digitales"},
		{[]string{"pedidosya", "rappi", "mcdonalds", "burger", "grido", "mostaza"}, "comida", "delivery"},
		{[]string{"metrogas", "aysa", "edenor", "edesur", "personal flow", "claro", "telecom"}, "servicios", "hogar"},
		{[]string{"intereses pagados", "mantenimiento"}, "financiero", "comisiones/intereses"},
		{[]string{"reintegro promoción", "devolucion"}, "ingresos", "reintegros"},
		{[]string{"sueldo", "haberes"}, "ingresos", "sueldo"},
	}

	for _, rule := range rules {
		for _, kw := range rule.keywords {
			if strings.Contains(desc, kw) {
				c, s := rule.cat, rule.sub
				return &c, &s
			}
		}
	}

	return nil, nil
}
