package utils

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
)

// GetBinary return bytes the binary
func GetBinary(urlBinary string) ([]byte, int, error) {
	client := &http.Client{
		Timeout: time.Second * 90,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlBinary, nil)
	if err != nil {
		return nil, 0, err
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}

	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}

	if res.StatusCode == http.StatusOK {
		return respBytes, res.StatusCode, nil
	}

	return nil, 0, pkg.ErrNotFound

}
