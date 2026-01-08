package processor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/juank/finance-ai/backend/internal/db"
	"github.com/juank/finance-ai/backend/internal/models"
	"github.com/juank/finance-ai/backend/internal/processor/common"
	"github.com/juank/finance-ai/backend/internal/processor/parsers"
)

type Engine struct {
	OutputDir string
	UserID    uuid.UUID
}

func NewEngine(outputDir string, userID uuid.UUID) *Engine {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		os.MkdirAll(outputDir, 0755)
	}
	return &Engine{OutputDir: outputDir, UserID: userID}
}

func (e *Engine) RunAll() error {
	// Re-scan hardcoded paths (Legacy or for dev)
	// For production, we prefer ProcessFile called from API
	var allFilesTransactions []models.Transaction

	// Simplified for brevity in this example, normally you'd map dirs to parsers
	dirs := []struct {
		path   string
		ext    string
		parser common.Normalizer
	}{
		{"/Users/juank/Documents/Cuentas/Bancos/Brubank", ".pdf", &parsers.BrubankPDFParser{}},
		{"/Users/juank/Documents/Cuentas/Bancos/MercadoPago", ".csv", &parsers.MercadoPagoParser{}},
		{"/Users/juank/Documents/Cuentas/Bancos/Deel", ".csv", &parsers.DeelParser{}},
	}

	for _, d := range dirs {
		txs, _ := e.processDir(d.path, d.ext, d.parser, uuid.Nil)
		allFilesTransactions = append(allFilesTransactions, txs...)
	}

	return e.SaveAndConsolidate(allFilesTransactions)
}

func (e *Engine) ProcessFile(filePath string, parser common.Normalizer, uploadID uuid.UUID) ([]models.Transaction, error) {
	txs, err := parser.Normalize(filePath)
	if err != nil {
		return nil, err
	}

	// Enrich with metadata
	for i := range txs {
		txs[i].UserID = e.UserID
		txs[i].UploadID = uploadID
		txs[i].ProcessedAt = time.Now()
	}

	return txs, nil
}

func (e *Engine) SaveAndConsolidate(txs []models.Transaction) error {
	// 5. Neutralization & Deduplication
	txs = deduplicate(txs)
	txs = NeutralizeTransfers(txs)

	// Persist to DB
	if err := db.GetDB().UpsertTransactions(txs); err != nil {
		return err
	}

	// 6. Sort and save consolidated JSON (Optional/Legacy support)
	sort.Slice(txs, func(i, j int) bool {
		return txs[i].Date > txs[j].Date
	})
	e.saveJSON("consolidated_transactions.json", txs)

	return nil
}

func (e *Engine) processDir(dir, ext string, parser common.Normalizer, uploadID uuid.UUID) ([]models.Transaction, error) {
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
			txs, err := e.ProcessFile(filepath.Join(dir, f.Name()), parser, uploadID)
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
