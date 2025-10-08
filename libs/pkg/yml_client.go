package pkg

import (
	"gopkg.in/yaml.v2"
)

// YmlClient is struct for yml functions
type YmlClient struct{}

// NewYmlClient return instance of yml client
func NewYmlClient() *YmlClient {
	return &YmlClient{}
}

// Unmarshall execute parse the yml for data
func (y *YmlClient) Unmarshall(data []byte, fileYml any) error {
	if err := yaml.Unmarshal(data, fileYml); err != nil {
		return err
	}
	return nil
}

// Marshall execute parse the data for yml
func (y *YmlClient) Marshall(fileYml any) ([]byte, error) {
	data, err := yaml.Marshal(fileYml)
	if err != nil {
		return nil, err
	}
	return data, nil
}
