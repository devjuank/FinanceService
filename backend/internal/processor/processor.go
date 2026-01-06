package processor

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

type Processor struct {
	PythonPath string
	ScriptsDir string
	OutputDir  string
}

func NewProcessor(pythonPath, scriptsDir, outputDir string) *Processor {
	return &Processor{
		PythonPath: pythonPath,
		ScriptsDir: scriptsDir,
		OutputDir:  outputDir,
	}
}

func (p *Processor) ProcessFile(filePath string, source string) error {
	// For now, we manually trigger the Python normalizer logic
	// In a real system, we might have a specific CLI for the normalizer
	// Since normalizer.py runs everything at once, we might want to adapt it

	cmd := exec.Command(p.PythonPath, filepath.Join(p.ScriptsDir, "normalizer.py"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python script failed: %v, output: %s", err, string(output))
	}

	// Then run the secondary processing
	cmd = exec.Command(p.PythonPath, filepath.Join(p.ScriptsDir, "process_transactions.py"))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("process_transactions script failed: %v, output: %s", err, string(output))
	}

	return nil
}
