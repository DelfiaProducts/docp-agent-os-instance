package api

import (
	"net/http"

	controllers "github.com/DelfiaProducts/docp-agent-os-instance/libs/controllers"
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/gorilla/mux"
)

// CommonRoutes is struct for routes the common
type CommonRoutes struct {
	controller *controllers.CommonHttpController
	logger     libinterfaces.ILogger
}

// NewCommonRoutes return instance of common routers
func NewCommonRoutes(logger libinterfaces.ILogger) *CommonRoutes {
	return &CommonRoutes{
		logger: logger,
	}
}

// Setup execute configuration
func (c *CommonRoutes) Setup() error {
	controller := controllers.NewCommonHttpController(c.logger)
	if err := controller.Setup(); err != nil {
		return err
	}
	c.controller = controller
	return nil
}

// Health is handler for health
func (c *CommonRoutes) Health(w http.ResponseWriter, r *http.Request) {
	c.controller.Health(w, r)
}

// BuildRoutes execute build the routes datadog
func (c *CommonRoutes) BuildRoutes(router *mux.Router) error {
	route := router.PathPrefix("/health").Subrouter()
	route.HandleFunc("", c.Health).Methods("GET")
	return nil
}
