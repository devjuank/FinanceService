package processor

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/juank/finance-ai/backend/internal/models"
	"github.com/juank/finance-ai/backend/internal/processor/common"
	"github.com/juank/finance-ai/backend/internal/processor/parsers"
)

type Engine struct {
	OutputDir string
}

func NewEngine(outputDir string) *Engine {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}
	return &Engine{OutputDir: outputDir}
}

func (e *Engine) RunAll() error {
	var allTransactions []models.Transaction

	// 1. Process Brubank
	brubankDir := "/Users/juank/Documents/Cuentas/Bancos/Brubank"
	brubankParser := &parsers.BrubankPDFParser{}
	brubankTXs, _ := e.processDir(brubankDir, ".pdf", brubankParser)
	if len(brubankTXs) > 0 {
		e.saveJSON("brubank.json", brubankTXs)
		allTransactions = append(allTransactions, brubankTXs...)
	}

	// 2. Process MercadoPago
	mpDir := "/Users/juank/Documents/Cuentas/Bancos/MercadoPago"
	mpParser := &parsers.MercadoPagoParser{}
	mpTXs, _ := e.processDir(mpDir, ".csv", mpParser)
	if len(mpTXs) > 0 {
		e.saveJSON("mercadopago.json", mpTXs)
		allTransactions = append(allTransactions, mpTXs...)
	}

	// 3. Process Deel
	deelDir := "/Users/juank/Documents/Cuentas/Bancos/Deel"
	deelParser := &parsers.DeelParser{}
	deelTXs, _ := e.processDir(deelDir, ".csv", deelParser)
	if len(deelTXs) > 0 {
		e.saveJSON("deel.json", deelTXs)
		allTransactions = append(allTransactions, deelTXs...)
	}

	// 4. Process Santander
	santanderDir := "/Users/juank/Documents/Cuentas/Bancos/santander"
	santanderXLSXParser := &parsers.SantanderXLSXParser{}
	santanderTXs, _ := e.processDir(santanderDir, ".xlsx", santanderXLSXParser)

	tarjetaDir := filepath.Join(santanderDir, "Tarjeta")
	visaParser := &parsers.SantanderVisaPDFParser{}
	visaTXs, _ := e.processDir(tarjetaDir, ".pdf", visaParser)

	santanderAll := append(santanderTXs, visaTXs...)
	if len(santanderAll) > 0 {
		e.saveJSON("santander.json", santanderAll)
		allTransactions = append(allTransactions, santanderAll...)
	}

	// 5. Neutralization & Deduplication
	// Deduplicate by ID
	allTransactions = deduplicate(allTransactions)

	// Neutralize transfers
	allTransactions = NeutralizeTransfers(allTransactions)

	// 6. Sort and save consolidated
	sort.Slice(allTransactions, func(i, j int) bool {
		return allTransactions[i].Date > allTransactions[j].Date
	})

	e.saveJSON("consolidated_transactions.json", allTransactions)

	fmt.Printf("Total processed transactions: %d\n", len(allTransactions))
	return nil
}

func (e *Engine) processDir(dir, ext string, parser common.Normalizer) ([]models.Transaction, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var all []models.Transaction
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ext {
			txs, err := parser.Normalize(filepath.Join(dir, f.Name()))
			if err == nil {
				all = append(all, txs...)
			}
		}
	}
	return all, nil
}

func (e *Engine) saveJSON(filename string, data interface{}) {
	path := filepath.Join(e.OutputDir, filename)
	file, _ := os.Create(path)
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

func deduplicate(txs []models.Transaction) []models.Transaction {
	seen := make(map[string]bool)
	var unique []models.Transaction
	for _, tx := range txs {
		if !seen[tx.ID] {
			seen[tx.ID] = true
			unique = append(unique, tx)
		}
	}
	return unique
}
