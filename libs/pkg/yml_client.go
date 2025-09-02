package pkg

import (
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"gopkg.in/yaml.v2"
)

// YmlClient is struct for yml functions
type YmlClient struct{}

// NewYmlClient return instance of yml client
func NewYmlClient() *YmlClient {
	return &YmlClient{}
}

// Unmarshall execute parse the yml for data
func (y *YmlClient) Unmarshall(data []byte, config *dto.ConfigAgent) error {
	if err := yaml.Unmarshal(data, config); err != nil {
		return err
	}
	return nil
}

// Marshall execute parse the data for yml
func (y *YmlClient) Marshall(config *dto.ConfigAgent) ([]byte, error) {
	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}
	return data, nil
}
