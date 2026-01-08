package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/juank/finance-ai/backend/internal/processor"
)

func main() {
	fmt.Println("Starting Financial Processor (Go Native)...")

	outputDir := "/Users/juank/Documents/Cuentas/DatosClasificados"
	engine := processor.NewEngine(outputDir, uuid.Nil)

	err := engine.RunAll()
	if err != nil {
		log.Fatalf("Error running processor: %v", err)
	}

	fmt.Println("Processing completed successfully.")
}
