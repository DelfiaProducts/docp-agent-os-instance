package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/services"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// parseParams parse params
func parseParams(apiKey, tags, version, noGroupAssociation, oryaSite, vmName *string) {
	flag.StringVar(apiKey, "api_key", "", "orya api key")
	flag.StringVar(tags, "tags", "", "orya tags")
	flag.StringVar(version, "version", "latest", "orya version")
	flag.StringVar(noGroupAssociation, "no_group_association", "false", "no group association")
	flag.StringVar(oryaSite, "orya_site", "", "orya site for services")
	flag.StringVar(vmName, "vm_name", "", "vm name")
	flag.Parse()
}

// prepareUrl returns the binary URL
func prepareUrl(url, version, fileName string) string {
	versionName := "latest"
	if len(version) > 0 {
		versionName = version
	}
	return fmt.Sprintf("%s/%s/%s", url, versionName, fileName)
}

// downloadFile downloads the binary from the bucket
func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("binary not found: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("error creating binary file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error copying binary to file: %v", err)
	}

	return nil
}

// notifyError displays an error in a Windows MessageBox
func notifyError(title, message string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	msgBox := user32.NewProc("MessageBoxW")

	titleUTF16, _ := syscall.UTF16PtrFromString(title)
	messageUTF16, _ := syscall.UTF16PtrFromString(message)

	msgBox.Call(0, uintptr(unsafe.Pointer(messageUTF16)), uintptr(unsafe.Pointer(titleUTF16)), 0x10)
	os.Exit(1)
}

// notifyUsageLimitExceeded shows a clear message for scenario 2
func notifyUsageLimitExceeded(limit int) {
	msg := fmt.Sprintf("Usage limit exceeded.\n\nYou have reached the maximum number of agents for your plan.\nCurrent limit: %d\n\nIncrease your limits on the Orya platform.\n", limit)
	notifyError("Orya Installer", msg)
}

// notifyNetworkError shows a message for scenario 3
func notifyNetworkError(details string) {
	msg := fmt.Sprintf("Could not connect to Orya.\n\nCheck:\n- If the Orya Site URL is correct\n- If your network connection is working\n- If the service is reachable from this machine\n\nTechnical details: %s", details)
	notifyError("Orya Installer", msg)
}

// notifyBackendError shows a message for scenario 4
func notifyBackendError() {
	notifyError("Orya Installer", "The Orya backend is temporarily unavailable.\n\nPlease wait a few moments and try again.\nIf the problem persists, contact support.")
}

// notifyLocalError shows a message for scenario 5 (local OS error)
func notifyLocalError(operation string, err error) {
	msg := fmt.Sprintf("Local system error during installation.\n\nOperation: %s\nError: %s\n\nCheck system requirements and try again.", operation, err.Error())
	notifyError("Orya Installer", msg)
}

// notifyAlreadyRunning shows a message when the agent is already running
func notifyAlreadyRunning() {
	notifyError("Orya Installer", "The Orya agent is already running.\n\nIf you need to reinstall, remove the current agent first.")
}

// isAgentRunning checks if the manager.exe process is already running on Windows
func isAgentRunning() bool {
	program := pkg.NewExecProgram()
	output, err := program.ExecuteWithOutput("tasklist", []string{}, "/NH", "/FI", "IMAGENAME eq manager.exe")
	if err != nil {
		return false
	}
	return strings.Contains(output, "manager.exe")
}

// Main function
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
		notifyLocalError("service setup", err)
	}

	// Verify agent is not already running (cheapest check — local, no network)
	fmt.Println("Checking if the agent is already running...")
	if isAgentRunning() {
		notifyAlreadyRunning()
	}
	fmt.Println("OK.")

	// Verify Orya Site (scenario 3 — DNS / connectivity)
	fmt.Println("Checking access to Orya Site...")
	if err := utilityService.ValidateOryaSite(oryaSite); err != nil {
		notifyNetworkError(err.Error())
	}
	fmt.Println("Orya Site reachable.")

	// Verify usage limit + API Key
	fmt.Println("Checking usage limit...")
	usageLimitResponse, err := utilityService.ValidateUsageLimit(apiKey, oryaSite)
	if err != nil {
		switch {
		case errors.Is(err, pkg.ErrApiKeyInvalid):
			notifyError("Orya Installer", "Invalid API Key or wrong Orya Site.\n\nCheck:\n- If your API Key is correct\n- If the Orya Site URL is correct")
		case errors.Is(err, pkg.ErrUsageLimitExceeded):
			notifyUsageLimitExceeded(usageLimitResponse.Limit)
		case errors.Is(err, pkg.ErrNetworkError):
			notifyNetworkError(err.Error())
		case errors.Is(err, pkg.ErrBackendError):
			notifyBackendError()
		default:
			notifyLocalError("usage limit check", err)
		}
	}
	fmt.Println("Usage limit OK.")

	// Step 3: Resolve version
	fmt.Println("Resolving version... ")
	if version == "latest" {
		fmt.Println("Fetching latest agent version...")
		agentVersions, err := utilityService.FetchAgentVersions()
		if err != nil {
			notifyLocalError("version resolution", err)
		}
		version = agentVersions.LatestVersion
	}

	// Step 4: Download binary
	fmt.Println("Preparing download...")
	url := prepareUrl(baseUrl, version, fileName)
	actualDirectory, err := os.Getwd()
	if err != nil {
		notifyLocalError("get working directory", err)
	}

	destFile := filepath.Join(actualDirectory, "install_windows.msi")

	fmt.Println("Downloading installer...")
	err = downloadFile(url, destFile)
	if err != nil {
		notifyLocalError("installer download", err)
	}

	// Step 5: Execute MSI installer
	fmt.Println("Running installer...")
	program := pkg.NewExecProgram()
	command := fmt.Sprintf(`start-process -Wait msiexec -ArgumentList '/qn /i "%s" VERSION="%s" API_KEY="%s" TAGS="%s" NO_GROUP_ASSOCIATION="%s" ORYA_SITE="%s" VM_NAME="%s"'`, destFile, version, apiKey, tags, noGroupAssociation, oryaSite, vmName)

	_, err = program.ExecuteWithOutput("powershell", []string{}, "-Command", command)
	if err != nil {
		notifyLocalError("MSI execution", err)
	}

	// Cleanup
	if err := os.Remove(destFile); err != nil {
		notifyLocalError("cleanup", err)
	}
	fmt.Println("Installation completed successfully.")
	os.Exit(0)
}
