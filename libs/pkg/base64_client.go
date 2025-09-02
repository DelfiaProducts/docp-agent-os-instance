package pkg

import "encoding/base64"

// Base64Client is struct for base64 functions
type Base64Client struct{}

// NewBase64Client return instance of base64 client
func NewBase64Client() *Base64Client {
	return &Base64Client{}
}

// Decode execute decode the base64 from content
func (b *Base64Client) Decode(base64Content string) ([]byte, error) {
	content, err := base64.StdEncoding.DecodeString(base64Content)
	if err != nil {
		return nil, err
	}
	return content, nil
}
