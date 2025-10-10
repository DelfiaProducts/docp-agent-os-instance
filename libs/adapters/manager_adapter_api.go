package adapters

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// OryaAgentApiInstallDatadog execute call to api orya for install datadog agent
func (l *ManagerAdapter) OryaAgentApiInstallDatadog(ddApiKey, ddSite, version string) ([]byte, error) {
	l.logger.Debug("execute send request for install datadog agent", "trace", "agent-os-instance.manager_adapter.OryaAgentApiInstallDatadog", "ddApiKey", ddApiKey, "ddSite", ddSite)

	transaction := utils.NewTransactionStatus()
	ctx := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("install_ORYA_vendor_received", pkg.TransactionEventOpen, "install orya vendor received", ctx)
	time.Sleep(l.delay)

	datadogDto := dto.DatadogInstallDTO{
		DDSite:   ddSite,
		DDApiKey: ddApiKey,
		Version:  version,
	}
	bDatadogDto, err := l.json.Marshall(&datadogDto)
	if err != nil {
		go l.NotifyStatus("install_ORYA_vendor_error", pkg.TransactionEventClose, "failed install orya vendor", ctx)
		return nil, err
	}
	go l.NotifyStatus("install_ORYA_vendor_processing", pkg.TransactionEventUpdate, "install orya vendor processing", ctx)
	time.Sleep(l.delay)

	urlOryaInstallDatadog := fmt.Sprintf("http://127.0.0.1:%s/datadog/install", l.oryaApiPort)
	respBytes, err := l.requestForAgentInstallDatadog(urlOryaInstallDatadog, http.MethodPost, bDatadogDto)
	if err != nil {
		go l.NotifyStatus("install_ORYA_vendor_error", pkg.TransactionEventClose, "failed install orya vendor", ctx)
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.OryaAgentApiInstallDatadog", "error", err.Error())
		return nil, err
	}

	go l.NotifyStatus("install_ORYA_vendor_completed", pkg.TransactionEventClose, "install orya vendor completed", ctx)
	return respBytes, nil
}

// OryaAgentApiInstallDatadog execute call to api orya for install datadog agent
func (l *ManagerAdapter) OryaAgentApiInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationLibraries string) ([]byte, error) {
	l.logger.Debug("execute send request for install datadog agent with apm single step", "trace", "agent-os-instance.manager_adapter.OryaAgentApiInstallDatadogWithApmSingleStep", "ddApiKey", ddApiKey, "ddSite", ddSite, "ddApmInstrumentationLibraries", ddApmInstrumentationLibraries)

	transaction := utils.NewTransactionStatus()
	ctx := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("install_ORYA_vendor_tracer_received", pkg.TransactionEventOpen, "install orya vendor tracer single step received", ctx)
	time.Sleep(l.delay)

	datadogDto := dto.DatadogInstallDTO{
		DDSite:                        ddSite,
		DDApiKey:                      ddApiKey,
		Component:                     "tracer",
		Mode:                          "single_step",
		DDApmInstrumentationLibraries: ddApmInstrumentationLibraries,
	}
	bDatadogDto, err := l.json.Marshall(&datadogDto)
	if err != nil {
		go l.NotifyStatus("install_ORYA_vendor_tracer_error", pkg.TransactionEventClose, "failed install orya vendor tracer single step", ctx)
		return nil, err
	}

	urlOryaInstallDatadog := fmt.Sprintf("http://127.0.0.1:%s/datadog/install", l.oryaApiPort)

	go l.NotifyStatus("install_ORYA_vendor_tracer_processing", pkg.TransactionEventUpdate, "install orya vendor tracer single step processing", ctx)
	time.Sleep(l.delay)

	respBytes, err := l.requestForAgentInstallDatadog(urlOryaInstallDatadog, http.MethodPost, bDatadogDto)
	if err != nil {
		go l.NotifyStatus("install_ORYA_vendor_tracer_error", pkg.TransactionEventClose, "failed install orya vendor tracer single step", ctx)
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.OryaAgentApiInstallDatadogWithApmSingleStep", "error", err.Error())
		return nil, err
	}

	if err := l.SaveAlreadyTracer(true); err != nil {
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.OryaAgentApiInstallDatadogWithApmSingleStep", "error", err.Error())
		go l.NotifyStatus("install_ORYA_vendor_tracer_error", pkg.TransactionEventClose, "failed install orya vendor tracer single step", ctx)
		return nil, err
	}

	go l.NotifyStatus("install_ORYA_vendor_tracer_completed", pkg.TransactionEventClose, "install orya vendor tracer single step completed", ctx)
	return respBytes, nil
}

// OryaAgentApiInstallDatadog execute call to api orya for install datadog agent
func (l *ManagerAdapter) OryaAgentApiInstallDatadogWithApmTracingLibrary(ddApiKey, ddSite, language, pathTracer, version string) ([]byte, error) {
	l.logger.Debug("execute send request for install datadog agent with apm single step", "trace", "agent-os-instance.manager_adapter.OryaAgentApiInstallDatadogWithApmSingleStep", "ddApiKey", ddApiKey, "ddSite", ddSite, "language", language, "pathTracer", pathTracer, "version", version)

	transaction := utils.NewTransactionStatus()
	ctx := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("install_ORYA_vendor_tracer_received", pkg.TransactionEventOpen, "install orya vendor tracer library received", ctx)
	time.Sleep(l.delay)

	datadogDto := dto.DatadogInstallDTO{
		DDSite:    ddSite,
		DDApiKey:  ddApiKey,
		Component: "tracer",
		Mode:      "tracing_library",
		EnvVars: []dto.DatadogEnvVars{
			{
				Name:  "language",
				Value: language,
			},
			{
				Name:  "path_tracer",
				Value: pathTracer,
			},
			{
				Name:  "version",
				Value: version,
			},
		},
	}
	bDatadogDto, err := l.json.Marshall(&datadogDto)
	if err != nil {
		go l.NotifyStatus("install_ORYA_vendor_tracer_error", pkg.TransactionEventClose, "failed install orya vendor tracer", ctx)
		return nil, err
	}
	urlOryaInstallDatadog := fmt.Sprintf("http://127.0.0.1:%s/datadog/tracer/install", l.oryaApiPort)

	go l.NotifyStatus("install_ORYA_vendor_tracer_processing", pkg.TransactionEventUpdate, "install orya vendor tracer library processing", ctx)
	time.Sleep(l.delay)

	respBytes, err := l.requestForAgentInstallDatadog(urlOryaInstallDatadog, http.MethodPost, bDatadogDto)
	if err != nil {
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.OryaAgentApiInstallDatadogWithApmSingleStep", "error", err.Error())
		go l.NotifyStatus("install_ORYA_vendor_tracer_error", pkg.TransactionEventClose, "failed install orya vendor tracer", ctx)
		return nil, err
	}
	if err := l.SaveAlreadyTracer(true); err != nil {
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.OryaAgentApiInstallDatadogWithApmSingleStep", "error", err.Error())
		go l.NotifyStatus("install_ORYA_vendor_tracer_error", pkg.TransactionEventClose, "failed install orya vendor tracer", ctx)
		return nil, err
	}

	go l.NotifyStatus("install_ORYA_vendor_tracer_completed", pkg.TransactionEventClose, "install orya vendor tracer library completed", ctx)
	return respBytes, nil
}

// OryaAgentApiUninstallDatadog execute call to api orya for uninstall datadog agent
func (l *ManagerAdapter) OryaAgentApiUninstallDatadog() ([]byte, error) {
	l.logger.Debug("execute send request for uninstall datadog agent", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUninstallDatadog")

	transaction := utils.NewTransactionStatus()
	ctxTransaction := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("uninstall_ORYA_vendor_received", pkg.TransactionEventOpen, "uninstall orya vendor received", ctxTransaction)
	time.Sleep(l.delay)

	urlOryaUninstallDatadog := fmt.Sprintf("http://127.0.0.1:%s/datadog/uninstall", l.oryaApiPort)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlOryaUninstallDatadog, nil)
	if err != nil {
		go l.NotifyStatus("uninstall_ORYA_vendor_error", pkg.TransactionEventClose, "failed uninstall vendor", ctxTransaction)
		l.logger.Error("error in create request", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUninstallDatadog", "error", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := l.client.Do(req)
	if err != nil {
		go l.NotifyStatus("uninstall_ORYA_vendor_error", pkg.TransactionEventClose, "failed uninstall vendor", ctxTransaction)
		l.logger.Error("error in execute request", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUninstallDatadog", "error", err.Error())
		return nil, err
	}

	go l.NotifyStatus("uninstall_ORYA_vendor_processing", pkg.TransactionEventUpdate, "uninstall orya vendor processing", ctxTransaction)
	time.Sleep(l.delay)

	go l.NotifyStatus("uninstall_ORYA_vendor_complete", pkg.TransactionEventClose, "uninstall orya vendor completed", ctxTransaction)
	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		go l.NotifyStatus("uninstall_ORYA_vendor_error", pkg.TransactionEventClose, "failed uninstall vendor", ctxTransaction)
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUninstallDatadog", "error", err.Error())
		return nil, err
	}
	if err := l.SaveAlreadyTracer(false); err != nil {
		go l.NotifyStatus("uninstall_ORYA_vendor_error", pkg.TransactionEventClose, "failed uninstall vendor", ctxTransaction)
		l.logger.Error("error in save already tracer", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUninstallDatadog", "error", err.Error())
		return nil, err
	}

	return respBytes, nil
}

// OryaAgentApiUpdateConfigurationsDatadog execute call to api orya for update datadog configurations
func (l *ManagerAdapter) OryaAgentApiUpdateConfigurationsDatadog(content []byte) ([]byte, error) {
	l.logger.Debug("execute send request for update configurations in datadog agent", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUpdateConfigurationsDatadog", "content", string(content))

	transaction := utils.NewTransactionStatus()
	ctxTransaction := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("update_vendor_received", pkg.TransactionEventOpen, "update vendor received", ctxTransaction)
	time.Sleep(l.delay)

	urlOryaUpdateConfigurations := fmt.Sprintf("http://127.0.0.1:%s/datadog/configurations", l.oryaApiPort)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlOryaUpdateConfigurations, bytes.NewBuffer(content))
	if err != nil {
		go l.NotifyStatus("update_ORYA_vendor_error", pkg.TransactionEventClose, "failed update datadog", ctxTransaction)
		l.logger.Error("error in create request", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUpdateConfigurationsDatadog", "error", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := l.client.Do(req)
	if err != nil {
		go l.NotifyStatus("update_ORYA_vendor_error", pkg.TransactionEventClose, "failed update datadog", ctxTransaction)
		l.logger.Error("error in execute request", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUpdateConfigurationsDatadog", "error", err.Error())
		return nil, err
	}

	go l.NotifyStatus("update_ORYA_vendor_processing", pkg.TransactionEventUpdate, "update orya vendor processing", ctxTransaction)
	time.Sleep(l.delay)

	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		go l.NotifyStatus("update_ORYA_vendor_error", pkg.TransactionEventClose, "failed update datadog", ctxTransaction)
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUpdateConfigurationsDatadog", "error", err.Error())
		return nil, err
	}

	go l.NotifyStatus("update_ORYA_vendor_complete", pkg.TransactionEventClose, "update orya vendor completed", ctxTransaction)
	return respBytes, nil
}

// OryaAgentApiUpdateVersionDatadog execute call to api orya for update datadog version
func (l *ManagerAdapter) OryaAgentApiUpdateVersionDatadog(version string) ([]byte, error) {
	l.logger.Debug("execute send request for update version in datadog agent", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUpdateVersionDatadog", "version", version)
	updateVersion := dto.DatadogUpdateVersionDTO{Version: version}
	bodyUpdateVersion, err := l.json.Marshall(updateVersion)
	if err != nil {
		return nil, err
	}

	urlOryaUpdateVersion := fmt.Sprintf("http://127.0.0.1:%s/datadog/update/version", l.oryaApiPort)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlOryaUpdateVersion, bytes.NewBuffer(bodyUpdateVersion))
	if err != nil {
		l.logger.Error("error in create request", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUpdateVersionDatadog", "error", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := l.client.Do(req)
	if err != nil {
		l.logger.Error("error in execute request", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUpdateVersionDatadog", "error", err.Error())
		return nil, err
	}

	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.OryaAgentApiUpdateVersionDatadog", "error", err.Error())
		return nil, err
	}

	return respBytes, nil
}
