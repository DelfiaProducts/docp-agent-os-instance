package utils

import (
	"fmt"
	"net/url"
	"strings"
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
