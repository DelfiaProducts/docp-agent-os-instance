package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

// GetDomainUrl return domain url
func GetDomainUrl() (string, error) {
	docpDomain := os.Getenv("DOCP_DOMAIN")
	if len(docpDomain) != 0 {
		return docpDomain, nil
	}
	return pkg.DOCP_DOMAIN, nil
}

// GetCollectInterval return duration from env collect interval
func GetCollectInterval() (time.Duration, error) {
	var interval time.Duration
	docp_collect_interval_env := os.Getenv("DOCP_COLLECT_INTERVAL")
	if len(docp_collect_interval_env) == 0 {
		return time.Duration(time.Hour * 24), nil
	}
	duration, err := time.ParseDuration(docp_collect_interval_env)
	if err != nil {
		return interval, err
	}
	return duration, nil
}

// GetConfigFilePath return path the config file from env
func GetConfigFilePath() (string, error) {
	docp_config_file_path_env := os.Getenv("DOCP_CONFIG_FILE_PATH")
	if len(docp_config_file_path_env) == 0 {
		if runtime.GOOS == "windows" {
			programFiles := os.Getenv("ProgramFiles")
			return filepath.Join(programFiles, "DocpAgent", "config.yml"), nil
		} else {

			return filepath.Join(string(filepath.Separator), "opt", "docp-agent", "config.yml"), nil
		}
	}
	return docp_config_file_path_env, nil
}

// GetLogFilePath return path the log file
func GetLogFilePath() (string, error) {
	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("ProgramFiles")
		return filepath.Join(programFiles, "DocpAgent", "logs", "log.txt"), nil
	} else {
		return filepath.Join(string(filepath.Separator), "opt", "docp-agent", "logs", "log.txt"), nil
	}
}

// GetWorkDirPath return path the work dir from env
func GetWorkDirPath() (string, error) {
	docp_workdir_path_env := os.Getenv("DOCP_WORKDIR_PATH")
	if len(docp_workdir_path_env) == 0 {
		if runtime.GOOS == "windows" {
			programFiles := os.Getenv("ProgramFiles")
			return filepath.Join(programFiles, "DocpAgent"), nil
		} else {
			return filepath.Join(string(filepath.Separator), "opt", "docp-agent"), nil
		}
	}
	return docp_workdir_path_env, nil
}

// GetDatadogBinaryAgentPath return path the binary agent datadog
func GetDatadogBinaryAgentPath() (string, error) {
	if runtime.GOOS == "windows" {
		programFiles := os.Getenv("ProgramFiles")
		return filepath.Join(programFiles, "Datadog", "Datadog Agent", "bin", "agent"), nil
	} else {
		return filepath.Join(string(filepath.Separator), "opt", "datadog-agent", "bin", "agent"), nil
	}
}

// GetPortAgentApi return port the api agent from env
func GetPortAgentApi() (string, error) {
	docp_agent_port := os.Getenv("DOCP_AGENT_PORT")
	if len(docp_agent_port) == 0 {
		return pkg.DOCP_AGENT_PORT, nil
	}
	return docp_agent_port, nil
}
