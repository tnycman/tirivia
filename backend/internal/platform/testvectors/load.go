package testvectors

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Dir returns protocol/testvectors, located relative to this source file.
func Dir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "protocol", "testvectors")
}

// OpenDir opens Dir() as an os.Root so access cannot escape it.
func OpenDir() (*os.Root, error) { return os.OpenRoot(Dir()) }

// Load decodes a suite file.
func Load(name string, v any) error {
	root, err := OpenDir()
	if err != nil {
		return fmt.Errorf("testvectors: open dir: %w", err)
	}
	defer func() { _ = root.Close() }()
	b, err := root.ReadFile(name)
	if err != nil {
		return fmt.Errorf("testvectors: read %s: %w", name, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("testvectors: decode %s: %w", name, err)
	}
	return nil
}
