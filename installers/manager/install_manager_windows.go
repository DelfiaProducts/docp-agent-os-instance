package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"gopkg.in/yaml.v2"
)

// ConfigAgent is struct for config file agent
type ConfigAgent struct {
	Version            string `yaml:"version"`
	NoGroupAssociation bool   `yaml:"no_group_association,omitempty"`
	Agent              Agent  `yaml:"agent"`
}

// Agent is struct for config file agent
type Agent struct {
	ApiKey string            `yaml:"apiKey"`
	Tags   map[string]string `yaml:"tags"`
}

// parseParams parse params
func parseParams(apiKey, tags, version, noGroupAssociation *string) {
	flag.StringVar(apiKey, "API_KEY", "", "docp api key")
	flag.StringVar(tags, "TAGS", "", "docp tags")
	flag.StringVar(version, "VERSION", "latest", "docp version")
	flag.StringVar(noGroupAssociation, "NO_GROUP_ASSOCIATION", "true", "no group association")
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

// createDefaultFiles execute create the tree files
func createDefaultFiles(homeDir string) error {
	basePath := filepath.Join(homeDir, "DocpAgent")

	filePathEnv := filepath.Join(basePath, "environments")
	outEnv, err := os.Create(filePathEnv)
	if err != nil {
		return err
	}
	defer outEnv.Close()

	filesState := []string{"current", "received"}
	for _, file := range filesState {
		filePath := filepath.Join(basePath, "state", file)
		out, err := os.Create(filePath)
		if err != nil {
			return err
		}
		out.Close()
	}
	return nil
}

// processTags process tags
func processTags(tags string) map[string]string {
	tagSlice := strings.Split(tags, ",")
	tagMap := make(map[string]string)

	// Adiciona espaço após ':' e insere as tags no mapa
	for _, tag := range tagSlice {
		parts := strings.Split(tag, ":")
		if len(parts) == 2 {
			tagMap[parts[0]] = parts[1]
		}
	}
	return tagMap
}

// formatTagsForYAML recebe um slice de tags e formata para a estrutura YAML desejada
func formatTagsForYAML(tags []string) string {
	var formattedTags []string
	for _, tag := range tags {
		formattedTags = append(formattedTags, fmt.Sprintf("%s", tag))
	}
	return strings.Join(formattedTags, "\n")
}

// addContentConfigYml save initial config yml
func addContentConfigYml(docpFilesPath, apiKey, tags, version, noGroupAssociation string) error {
	processedTags := processTags(tags)
	noGroupBool, err := strconv.ParseBool(noGroupAssociation)
	if err != nil {
		return err
	}
	config := ConfigAgent{
		Version:            version,
		NoGroupAssociation: noGroupBool,
		Agent: Agent{
			ApiKey: apiKey,
			Tags:   processedTags,
		},
	}
	configFilePath := filepath.Join(docpFilesPath, "config.yml")

	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
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
	parseParams(&apiKey, &tags, &version, &noGroupAssociation)
	baseUrl := "https://github.com/DelfiaProducts/docp-agent-os-instance/releases/download"
	fileName := "manager-windows-amd64.exe"
	url := prepareUrl(baseUrl, version, fileName)

	pathDir := os.Getenv("ProgramFiles")
	docpFilesPath := filepath.Join(pathDir, "DocpAgent")
	if err := createDefaultFiles(pathDir); err != nil {
		notifyError("Installer Docp Manager", err.Error())
	}
	if err := addContentConfigYml(docpFilesPath, apiKey, tags, version, noGroupAssociation); err != nil {
		notifyError("Installer Docp Manager", err.Error())
	}

	destDir := filepath.Join(docpFilesPath, "bin")
	destFile := filepath.Join(destDir, "manager.exe")

	err := downloadFile(url, destFile)
	if err != nil {
		notifyError("Installer Docp Manager", err.Error())
	}
}
