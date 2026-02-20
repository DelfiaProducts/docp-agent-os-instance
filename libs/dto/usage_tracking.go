package dto

// UsageLimitResponse is struct for response the usage limit
// intalled agents
type UsageLimitResponse struct {
	HasLimit        bool `json:"has_limit"`
	CurrentUsage    int  `json:"current_usage"`
	Limit           int  `json:"limit"`
	ConfiguredLimit bool `json:"configured_limit"`
}
