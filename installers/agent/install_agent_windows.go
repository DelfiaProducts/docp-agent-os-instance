package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/services"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

// parseParams parse params
func parseParams(version *string) {
	flag.StringVar(version, "VERSION", "latest", "docp version")
	flag.Parse()
}

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

// notifyError execute notify error to windows
func notifyError(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	msgBox := user32.NewProc("MessageBoxW")

	titleUTF16, _ := syscall.UTF16PtrFromString(title)
	messageUTF16, _ := syscall.UTF16PtrFromString(message)

	msgBox.Call(0, uintptr(unsafe.Pointer(messageUTF16)), uintptr(unsafe.Pointer(titleUTF16)), 0x10)
	os.Exit(1)
}

// Função principal
func main() {
	var version string
	parseParams(&version)
	baseUrl := "https://github.com/DelfiaProducts/docp-agent-os-instance/releases/download"
	fileName := "agent-windows-amd64.exe"
	//verify if version latest
	if version == "latest" {
		logger := utils.NewDocpLoggerText(os.Stdout)
		utilityService := services.NewUtilityService(logger)
		if err := utilityService.Setup(); err != nil {
			notifyError("Installer Docp Agent", err.Error())
		}
		agentVersions, err := utilityService.FetchAgentVersions()
		if err != nil {
			notifyError("Installer Docp Agent", err.Error())
		}
		version = agentVersions.LatestVersion
	}
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
