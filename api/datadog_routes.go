package api

import (
	"net/http"

	controllers "github.com/DelfiaProducts/docp-agent-os-instance/libs/controllers"
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/gorilla/mux"
)

// DatadogRoutes is struct for routes the datadog
type DatadogRoutes struct {
	controller *controllers.DatadogHttpController
	logger     libinterfaces.ILogger
}

// NewDatadogRoutes return instance of datadog routers
func NewDatadogRoutes(logger libinterfaces.ILogger) *DatadogRoutes {
	return &DatadogRoutes{
		logger: logger,
	}
}

// Setup execute configuration
func (d *DatadogRoutes) Setup() error {
	controller := controllers.NewDatadogHttpController(d.logger)
	if err := controller.Setup(); err != nil {
		return err
	}
	d.controller = controller
	return nil
}

// InstallAgent is handler for install agent datadog
func (d *DatadogRoutes) InstallAgent(w http.ResponseWriter, r *http.Request) {
	d.controller.InstallAgent(w, r)
}

// InstallTracer is handler for install tracer datadog
func (d *DatadogRoutes) InstallTracer(w http.ResponseWriter, r *http.Request) {
	d.controller.InstallTracer(w, r)
}

// UninstallAgent is handler for uninstall agent datadog
func (d *DatadogRoutes) UninstallAgent(w http.ResponseWriter, r *http.Request) {
	d.controller.UninstallAgent(w, r)
}

// UpdateAgentConfigurations is handler for update agent configurations in datadog
func (d *DatadogRoutes) UpdateAgentConfigurations(w http.ResponseWriter, r *http.Request) {
	d.controller.UpdateAgentConfigurations(w, r)
}

// UpdateAgentVersion is handler for update agent version in datadog
func (d *DatadogRoutes) UpdateAgentVersion(w http.ResponseWriter, r *http.Request) {
	d.controller.UpdateAgentVersion(w, r)
}

// BuildRoutes execute build the routes datadog
func (d *DatadogRoutes) BuildRoutes(router *mux.Router) error {
	route := router.PathPrefix("/datadog").Subrouter()
	route.HandleFunc("/install", d.InstallAgent).Methods("POST")
	route.HandleFunc("/uninstall", d.UninstallAgent).Methods("POST")
	route.HandleFunc("/tracer/install", d.InstallTracer).Methods("POST")
	route.HandleFunc("/configurations", d.UpdateAgentConfigurations).Methods("POST")
	route.HandleFunc("/update/version", d.UpdateAgentVersion).Methods("POST")
	return nil
}
