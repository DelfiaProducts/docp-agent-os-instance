package utils

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
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
