package dto

// AuthPayload is struct for auth payload
type AuthPayload struct {
	ApiKey    string `json:"api_key"`
	ComputeId string `json:"compute_id"`
}

// AuthResponse is struct for auth response
type AuthResponse struct {
	AccessToken string `json:"access_token"`
}
