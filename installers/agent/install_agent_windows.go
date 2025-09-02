package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// prepareUrlAgent return url the binary
func prepareUrlAgent(url, version, fileName string) string {
	versionName := "latest"
	if len(version) > 0 {
		versionName = version
	}
	return fmt.Sprintf("%s/%s/%s", url, versionName, fileName)
}

// downloadFileAgent get binary file from bucket
func downloadFileAgent(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("binary not found: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error on create file for binary: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error on copy to file binary: %v", err)
	}

	return nil
}

// Função principal
func main() {
	var version string
	baseUrl := "https://test-docp-agent-data.s3.amazonaws.com/agent"
	fileName := "win_amd64.exe"
	url := prepareUrlAgent(baseUrl, version, fileName)

	pathDir := os.Getenv("ProgramFiles")
	docpFilesPath := filepath.Join(pathDir, "DocpAgent")

	destDir := filepath.Join(docpFilesPath, "bin")
	destFile := filepath.Join(destDir, "agent.exe")

	err := downloadFileAgent(url, destFile)
	if err != nil {
		panic(err)
	}
}
