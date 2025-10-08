package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
)

// OryaApi is struct for api orya
type OryaApi struct {
	port   string
	router *mux.Router
	srv    *http.Server
	logger libinterfaces.ILogger
}

// NewOryaApi return instance of orya api
func NewOryaApi(port string, logger libinterfaces.ILogger) *OryaApi {
	return &OryaApi{
		port:   port,
		logger: logger,
	}
}

// setupCommonRoutes execute configuration the common routes
func (d *OryaApi) setupCommonRoutes() error {
	commonRoutes := NewCommonRoutes(d.logger)
	if err := commonRoutes.Setup(); err != nil {
		return err
	}
	if err := commonRoutes.BuildRoutes(d.router); err != nil {
		return err
	}
	return nil
}

// setupDatadogRoutes execute configuration the routes datadog
func (d *OryaApi) setupDatadogRoutes() error {
	datadogRoutes := NewDatadogRoutes(d.logger)
	if err := datadogRoutes.Setup(); err != nil {
		return err
	}
	if err := datadogRoutes.BuildRoutes(d.router); err != nil {
		return err
	}
	return nil
}

// Setup execute configuration for api
func (d *OryaApi) Setup() error {
	router := mux.NewRouter()
	d.router = router
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", d.port),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}
	d.srv = srv
	if err := d.setupCommonRoutes(); err != nil {
		return err
	}
	if err := d.setupDatadogRoutes(); err != nil {
		return err
	}
	return nil
}

// Run execute running the api
func (d *OryaApi) Run() error {
	if err := d.srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}
