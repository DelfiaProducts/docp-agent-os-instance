package controllers

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/adapters"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
)

// DatadogHttpController is struct the datadog controller for
// api docp
type DatadogHttpController struct {
	logger  interfaces.ILogger
	adapter *adapters.DatadogAdapter
}

// NewDatadogHttpController return instance of datadog http controller
func NewDatadogHttpController(logger interfaces.ILogger) *DatadogHttpController {
	return &DatadogHttpController{
		logger: logger,
	}
}

// Setup execute configuration
func (d *DatadogHttpController) Setup() error {
	adapter := adapters.NewDatadogAdapter(d.logger)
	if err := adapter.Setup(); err != nil {
		return err
	}
	d.adapter = adapter
	return nil
}

// InstallTracer execute install the tracer datadog
func (d *DatadogHttpController) InstallTracer(w http.ResponseWriter, r *http.Request) {
	d.logger.Debug("install tracer", "trace", "docp-agent-os-instance.datadog_http_controller.InstallTracer")
	var datadogInstallDto dto.DatadogInstallDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&datadogInstallDto); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if errMarshal := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "error", Code: "DATADOG_INSTALL_ERR", Message: err.Error()}); errMarshal != nil {
			d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.InstallTracer", "error", errMarshal.Error())
			return
		}
	}
	if datadogInstallDto.Component == "tracer" {
		if datadogInstallDto.Mode == "tracing_library" {
			d.logger.Debug("install tracer", "trace", "docp-agent-os-instance.datadog_http_controller.InstallTracer", "datadogInstallDto", datadogInstallDto)
			language, pathTracer, version := d.adapter.GetApmEnvVarsTracingLibrary(datadogInstallDto.EnvVars)
			go d.adapter.InstallAgentApmTracingLibrary(language, pathTracer, version)
		}
	}
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "accepted", Code: "DATADOG_INSTALL_TRACER_ACCEPTED", Message: "accepted install"}); err != nil {
		d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.InstallTracer", "error", err.Error())
		return
	}
}

// InstallAgent execute install the agent datadog
func (d *DatadogHttpController) InstallAgent(w http.ResponseWriter, r *http.Request) {
	d.logger.Debug("install agent", "trace", "docp-agent-os-instance.datadog_http_controller.InstallAgent")
	var datadogInstallDto dto.DatadogInstallDTO
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&datadogInstallDto); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if errMarshal := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "error", Code: "DATADOG_INSTALL_ERR", Message: err.Error()}); errMarshal != nil {
			d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.InstallAgent", "error", errMarshal.Error())
			return
		}
	}
	if datadogInstallDto.Component == "tracer" {
		if datadogInstallDto.Mode == "single_step" {
			d.logger.Debug("install agent", "trace", "docp-agent-os-instance.datadog_http_controller.InstallAgent", "datadogInstallDto - single_step", datadogInstallDto)
			go d.adapter.InstallAgentApmSingleStep(datadogInstallDto.DDSite, datadogInstallDto.DDApiKey, datadogInstallDto.EnvVars)
		}
		if datadogInstallDto.Mode == "tracing_library" {
			d.logger.Debug("install agent", "trace", "docp-agent-os-instance.datadog_http_controller.InstallAgent", "datadogInstallDto - tracing_library", datadogInstallDto)
			language, pathTracer, version := d.adapter.GetApmEnvVarsTracingLibrary(datadogInstallDto.EnvVars)
			go d.adapter.InstallAgentApmTracingLibrary(language, pathTracer, version)
		}
	} else {
		d.logger.Debug("install agent", "trace", "docp-agent-os-instance.datadog_http_controller.InstallAgent", "datadogInstallDto", datadogInstallDto)
		go d.adapter.InstallAgent(datadogInstallDto.DDSite, datadogInstallDto.DDApiKey)
	}
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "accepted", Code: "DATADOG_INSTALL_ACCEPTED", Message: "accepted install"}); err != nil {
		d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.InstallAgent", "error", err.Error())
		return
	}
}

// UninstallAgent execute uninstall agent datadog
func (d *DatadogHttpController) UninstallAgent(w http.ResponseWriter, r *http.Request) {
	d.logger.Debug("uninstall agent", "trace", "docp-agent-os-instance.datadog_http_controller.UninstallAgent")
	go d.adapter.DPKGConfigure()
	time.Sleep(5 * time.Second)
	go d.adapter.UninstallAgent()
	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "accepted", Code: "DATADOG_UNINSTALL_ACCEPTED", Message: "accepted uninstall"}); err != nil {
		d.logger.Error("error in marshal response datadog", "unistall", "docp-agent-os-instance.datadog_http_controller.UninstallAgent", "error", err.Error())
		return
	}
}

// UpdateAgentConfigurations execute update the agent configurations datadog
func (d *DatadogHttpController) UpdateAgentConfigurations(w http.ResponseWriter, r *http.Request) {
	d.logger.Debug("update agent configurations", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentConfigurations")
	var datadogActionsFile dto.StateActionFiles

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&datadogActionsFile); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if errMarshal := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "error", Code: "DATADOG_UPDATE_CONFIGURATION_ERR", Message: err.Error()}); errMarshal != nil {
			d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentConfigurations", "error", errMarshal.Error())
			return
		}
		return
	}

	datadogFilePath, err := d.adapter.DiscoverDatadogConfigPath()
	d.logger.Debug("update agent configurations", "datadogFilePath", datadogFilePath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if errMarshal := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "error", Code: "DATADOG_UPDATE_CONFIGURATION_ERR", Message: err.Error()}); errMarshal != nil {
			d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentConfigurations", "error", errMarshal.Error())
			return
		}
		return
	}
	configPath := filepath.Join(datadogFilePath, datadogActionsFile.FilePath)
	d.logger.Debug("update agent configurations", "configPath", configPath)
	if err := d.adapter.BackupConfigFileDatadog(configPath, []byte(datadogActionsFile.Content)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if errMarshal := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "error", Code: "DATADOG_UPDATE_CONFIGURATION_ERR", Message: err.Error()}); errMarshal != nil {
			d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentConfigurations", "error", errMarshal.Error())
			return
		}
		return
	}

	if err := d.adapter.UpdateConfigFileDatadog(configPath); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if errMarshal := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "error", Code: "DATADOG_UPDATE_CONFIGURATION_ERR", Message: err.Error()}); errMarshal != nil {
			d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentConfigurations", "error", errMarshal.Error())
			return
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "accepted", Code: "DATADOG_UPDATED_CONFIGURATION_ACCEPTED", Message: "accepted update configurations"}); err != nil {
		d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentConfigurations", "error", err.Error())
		return
	}
}

// UpdateAgentVersion execute update the agent version datadog
func (d *DatadogHttpController) UpdateAgentVersion(w http.ResponseWriter, r *http.Request) {
	d.logger.Debug("update agent version", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentVersion")
	var datadogUpdateVersionDto dto.DatadogUpdateVersionDTO

	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&datadogUpdateVersionDto); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if errMarshal := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "error", Code: "DATADOG_UPDATE_VERSION_ERR", Message: err.Error()}); errMarshal != nil {
			d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentVersion", "error", errMarshal.Error())
			return
		}
		return
	}
	//update repository local
	if err := d.adapter.UpdateRepository(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if errMarshal := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "error", Code: "DATADOG_UPDATE_VERSION_ERR", Message: err.Error()}); errMarshal != nil {
			d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentVersion", "error", errMarshal.Error())
			return
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
	if err := json.NewEncoder(w).Encode(&dto.DatadogResponse{Status: "accepted", Code: "DATADOG_UPDATED_VERSION_ACCEPTED", Message: "accepted update version"}); err != nil {
		d.logger.Error("error in marshal response datadog", "trace", "docp-agent-os-instance.datadog_http_controller.UpdateAgentVersion", "error", err.Error())
		return
	}
}
