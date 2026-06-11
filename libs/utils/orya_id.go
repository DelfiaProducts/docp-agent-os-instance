package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
)

var (
	oryaIDOnce  sync.Once
	oryaIDValue string
	oryaIDErr   error
)

// GetOrCreateOryaID returns the persistent orya_id for this host.
// The ID is generated once via ULID and stored in a file outside the
// agent working directory, so it survives uninstall/reinstall.
//
// The path can be overridden via the ORYA_ID_FILE_PATH environment variable.
func GetOrCreateOryaID() (string, error) {
	oryaIDOnce.Do(func() {
		oryaIDValue, oryaIDErr = loadOrCreateOryaID()
	})
	return oryaIDValue, oryaIDErr
}

func loadOrCreateOryaID() (string, error) {
	filePath := getOryaIDFilePath()

	// Try to read existing orya_id from file
	id, err := readIDFromFile(filePath)
	if err == nil {
		return id, nil
	}

	// Generate a new ULID and normalize to lowercase
	id = strings.ToLower(ulid.Make().String())

	// Ensure parent directory exists, then write the ID
	if err := writeIDToFile(filePath, id); err != nil {
		return id, fmt.Errorf("failed to persist orya_id: %w", err)
	}

	return id, nil
}

// readIDFromFile reads a trimmed non-empty ID from the given file path.
func readIDFromFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", fmt.Errorf("empty orya_id file: %s", path)
	}
	return trimmed, nil
}

// writeIDToFile ensures the parent directory exists and writes the ID to the file.
func writeIDToFile(path, id string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	if err := os.WriteFile(path, []byte(id+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write to %s: %w", path, err)
	}
	return nil
}

func getOryaIDFilePath() string {
	if envPath := os.Getenv("ORYA_ID_FILE_PATH"); envPath != "" {
		return envPath
	}
	if runtime.GOOS == "windows" {
		programData := os.Getenv("ProgramData")
		if programData == "" {
			programData = `C:\ProgramData`
		}
		return filepath.Join(programData, "OryaAgent", "orya-id")
	}
	return filepath.Join(string(filepath.Separator), "opt", "orya-id")
}
