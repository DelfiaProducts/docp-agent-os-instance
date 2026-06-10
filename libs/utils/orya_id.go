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
	data, err := os.ReadFile(filePath)
	if err == nil && len(data) > 0 {
		return strings.TrimSpace(string(data)), nil
	}

	// Generate a new ULID and normalize to lowercase
	id := strings.ToLower(ulid.Make().String())

	// Ensure parent directory exists (important for Windows %ProgramData%)
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return id, fmt.Errorf("failed to create orya_id directory %s: %w", dir, err)
	}

	// Write the ID to the file
	if err := os.WriteFile(filePath, []byte(id+"\n"), 0644); err != nil {
		return id, fmt.Errorf("failed to write orya_id to %s: %w", filePath, err)
	}

	return id, nil
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
