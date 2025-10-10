package services

import (
	"path/filepath"
)

// getContentConfigFile return content of config file
func (s *StateCheckService) getContentConfigFile() ([]byte, error) {
	s.logger.Debug("get content config file", "trace", "agent-os-instance.state_check_service.getContentConfigFile")
	content, err := s.fileSystem.GetFileContent(filepath.Join(s.workDirPath, "config.yml"))
	if err != nil {
		return nil, err
	}
	return content, nil
}
