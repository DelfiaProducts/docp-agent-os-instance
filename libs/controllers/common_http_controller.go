package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// CommonHttpController is struct the common controller for
// api docp
type CommonHttpController struct {
	logger interfaces.ILogger
}

// NewCommonHttpController return instance of common http controller
func NewCommonHttpController(logger interfaces.ILogger) *CommonHttpController {
	return &CommonHttpController{
		logger: logger,
	}
}

// Setup execute configuration the controller
func (c *CommonHttpController) Setup() error {
	return nil
}

// Health execute verify if api is running
func (c *CommonHttpController) Health(w http.ResponseWriter, r *http.Request) {
	c.logger.Debug("health", "trace", "docp-agent-os-instance.common_http_controller.Health")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "success", Code: "HEALTH_OK", Message: "api is running"}); err != nil {
		c.logger.Error("error in marshal response health", "trace", "docp-agent-os-instance.common_http_controller.Health", "error", err.Error())
		return
	}
}
