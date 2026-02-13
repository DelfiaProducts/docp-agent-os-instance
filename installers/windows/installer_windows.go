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

	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/services"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// parseParams parse params
func parseParams(apiKey, tags, version, noGroupAssociation, oryaSite, vmName *string) {
	flag.StringVar(apiKey, "API_KEY", "", "orya api key")
	flag.StringVar(tags, "TAGS", "", "orya tags")
	flag.StringVar(version, "VERSION", "latest", "orya version")
	flag.StringVar(noGroupAssociation, "NO_GROUP_ASSOCIATION", "false", "no group association")
	flag.StringVar(oryaSite, "ORYA_SITE", "", "orya site for services")
	flag.StringVar(vmName, "VM_NAME", "", "vm name")
	flag.Parse()
}

// prepareUrl return url the binary
func prepareUrl(url, version, fileName string) string {
	versionName := "latest"
	if len(version) > 0 {
		versionName = version
	}
	return fmt.Sprintf("%s/%s/%s", url, versionName, fileName)
}

// downloadFile get binary file from bucket
func downloadFile(url, dest string) error {
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
	var apiKey string
	var tags string
	var version string
	var noGroupAssociation string
	var oryaSite string
	var vmName string
	parseParams(&apiKey, &tags, &version, &noGroupAssociation, &oryaSite, &vmName)
	if len(oryaSite) == 0 {
		oryaSite = "https://msapi.orya.tech"
	}
	baseUrl := "https://github.com/OryaHub/agent-os-instance/releases/download"
	fileName := "install_manager_windows.msi"

	logger := utils.NewOryaLoggerText(os.Stdout)
	utilityService := services.NewUtilityService(logger)
	if err := utilityService.Setup(); err != nil {
		notifyError("Installer Windows", err.Error())
	}
	//verify usage limit
	usageLimitResponse, err := utilityService.ValidateUsageLimit(apiKey, oryaSite)
	if err != nil {
		notifyError("Installer Windows", err.Error())
	}
	if !usageLimitResponse.HasLimit {
		notifyError("Installer Windows", fmt.Sprintf("usage limit exceeded: cannot install the agent.  actual usage/limit: %d/%d", usageLimitResponse.CurrentUsage, usageLimitResponse.Limit))
	}
	//verify if version latest
	if version == "latest" {
		agentVersions, err := utilityService.FetchAgentVersions()
		if err != nil {
			notifyError("Installer Windows", err.Error())
		}
		version = agentVersions.LatestVersion
	}
	url := prepareUrl(baseUrl, version, fileName)

	actualDirectory, err := os.Getwd()
	if err != nil {
		notifyError("Installer Windows", err.Error())
	}

	destFile := filepath.Join(actualDirectory, "install_windows.msi")

	err = downloadFile(url, destFile)
	if err != nil {
		notifyError("Installer Windows", err.Error())
	}

	program := pkg.NewExecProgram()
	command := fmt.Sprintf(`start-process -Wait msiexec -ArgumentList '/qn /i "%s" VERSION="%s" API_KEY="%s" TAGS="%s" NO_GROUP_ASSOCIATION="%s" ORYA_SITE="%s" VM_NAME="%s"'`, destFile, version, apiKey, tags, noGroupAssociation, oryaSite, vmName)

	_, err = program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
	if err != nil {
		notifyError("Installer Windows", err.Error())
	}
	if err := os.Remove(destFile); err != nil {
		notifyError("Installer Windows", err.Error())
	}
	os.Exit(0)
}
