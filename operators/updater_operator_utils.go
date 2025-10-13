package operators

import "github.com/OryaHub/agent-os-instance/libs/utils"

// getVersionFromReceived extracts the version from the received content
func (l *UpdaterOperator) getVersionFromReceived() (string, error) {
	var applyVersion string
	received, err := l.adapter.GetContentReceived()
	if err != nil {
		return applyVersion, err
	}
	l.logger.Info("received content", "content", string(received))
	if len(received) > 0 {
		version, err := l.adapter.GetAgentVersionFromSignal(received)
		if err != nil {
			return applyVersion, err
		}
		l.logger.Info("version received", "version", version)
		if len(version) > 0 {
			if version == "latest" {
				l.logger.Info("latest version received, update needed")
				agentVersions, err := l.adapter.FetchAgentVersions()
				if err != nil {
					l.logger.Error("error fetching agent versions", "error", err.Error())
					return applyVersion, err
				}
				applyVersion = agentVersions.LatestVersion

			} else {
				applyVersion = version
			}

		}
	}
	if len(applyVersion) == 0 {
		return applyVersion, utils.ErrAgentVersionNotFound()
	}

	return applyVersion, nil
}
