package utils

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/OryaHub/agent-os-instance/libs/pkg"
)

// TransformMapToSlice return slice of string from map
func TransformMapToSlice(mapp map[string]interface{}) ([]string, error) {
	slc := []string{}
	for key, value := range mapp {
		if value == nil {
			slc = append(slc, fmt.Sprintf("%s", key))
			continue
		}
		slc = append(slc, fmt.Sprintf("%s:%v", key, value))
	}
	return slc, nil
}

// GetBaseUrlSite return base url from site
func GetBaseUrlSite(site string) (string, error) {
	u, err := url.Parse(site)
	if err != nil {
		return "", err
	}

	host := u.Host

	parts := strings.Split(host, ".")
	if len(parts) >= 2 {
		host = strings.Join(parts[len(parts)-2:], ".")
	}
	return host, nil
}

// ChoiceInstallerOrUninstaller return action for commands
func ChoiceInstallerOrUninstaller(system, mode, action, version string) string {
	switch {
	//linux
	case system == "linux" && mode == "agent" && action == "install":
		release := fmt.Sprintf("%s/%s/install_agent_linux.sh", pkg.URL_RELEASE, version)
		return fmt.Sprintf("curl -L %s | bash", release)
	case system == "linux" && mode == "agent" && action == "uninstall":
		release := fmt.Sprintf("%s/%s/uninstall_agent_linux.sh", pkg.URL_RELEASE, version)
		return fmt.Sprintf("curl -L %s | bash", release)
	case system == "linux" && mode == "updater" && action == "install":
		release := fmt.Sprintf("%s/%s/install_updater_linux.sh", pkg.URL_RELEASE, version)
		return fmt.Sprintf("curl -L %s | bash", release)
	case system == "linux" && mode == "updater" && action == "uninstall":
		release := fmt.Sprintf("%s/%s/uninstall_updater_linux.sh", pkg.URL_RELEASE, version)
		return fmt.Sprintf("curl -L %s | bash", release)
	case system == "linux" && mode == "manager" && action == "uninstall":
		release := fmt.Sprintf("%s/%s/uninstall_manager_linux.sh", pkg.URL_RELEASE, version)
		return fmt.Sprintf("curl -L %s | bash", release)
	//windows
	case system == "windows" && mode == "manager" && action == "install":
		release := fmt.Sprintf("%s/%s/install_manager_windows.msi", pkg.URL_RELEASE, version)
		return fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s'`, release)
	case system == "windows" && mode == "agent" && action == "install":
		release := fmt.Sprintf("%s/%s/install_agent_windows.msi", pkg.URL_RELEASE, version)
		return fmt.Sprintf(`Start-Process -Wait msiexec -ArgumentList '/qn /i %s'`, release)
	//macos
	case system == "macos" && mode == "agent" && action == "install":
		release := fmt.Sprintf("%s/%s/install_agent_macos.sh", pkg.URL_RELEASE, version)
		return fmt.Sprintf("curl -L %s | bash", release)
	case system == "macos" && mode == "agent" && action == "uninstall":
		release := fmt.Sprintf("%s/%s/uninstall_agent_macos.sh", pkg.URL_RELEASE, version)
		return fmt.Sprintf("curl -L %s | bash", release)
	case system == "macos" && mode == "manager" && action == "uninstall":
		release := fmt.Sprintf("%s/%s/uninstall_manager_macos.sh", pkg.URL_RELEASE, version)
		return fmt.Sprintf("curl -L %s | bash", release)
	}
	return ""
}

// ChoiceMsiWindowsInstallerFile return file msi windows installer
func ChoiceMsiWindowsInstallerFile(version string) string {
	switch version {
	case "latest":
		return "datadog-agent-7-latest.amd64.msi"
	default:
		return fmt.Sprintf("datadog-agent-%s-1.x86_64.msi", version)
	}
}

// IsVersionGreater compare if versionA is greater than versionB
func IsVersionGreater(versionA string, versionB string) (bool, error) {
	partsA := strings.Split(versionA, ".")
	partsB := strings.Split(versionB, ".")

	if len(partsA) != 3 || len(partsB) != 3 {
		return false, ErrDatadogVersionInvalidFormat()
	}

	// 2. Itera sobre as partes e compara
	for i := 0; i < 3; i++ {
		// Converte a string da parte em um número inteiro (int)
		numA, err := strconv.Atoi(partsA[i])
		if err != nil {
			return false, ErrDatadogVersionInvalidFormat()
		}

		numB, err := strconv.Atoi(partsB[i])
		if err != nil {
			return false, ErrDatadogVersionInvalidFormat()
		}

		// Compara os números
		if numA > numB {
			return true, nil
		}
		if numA < numB {
			return false, nil
		}
	}

	return false, nil
}
