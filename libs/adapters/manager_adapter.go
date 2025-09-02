package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/components"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/services"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

// ManagerAdapter is struct for manager adapter
type ManagerAdapter struct {
	interval                 time.Duration
	hostStats                *pkg.HostStats
	chanMetadata             chan []byte
	chanClose                chan struct{}
	isClosed                 bool
	wg                       *sync.WaitGroup
	logger                   interfaces.ILogger
	program                  *pkg.ExecProgram
	osOperation              interfaces.IOSOperation
	fileSystem               *pkg.FileSystem
	ymlClient                *pkg.YmlClient
	client                   *http.Client
	agentWorkDir             string
	store                    *utils.Store
	stateCheck               *services.StateCheckService
	auth                     *services.AuthService
	utilityService           *services.UtilityService
	docpApiPort              string
	delay                    time.Duration
	pendingTransactionEvents []dto.TransactionStatus
	LockedEvents             bool
}

// NewManagerAdapter return instance of linux manager adapter
func NewManagerAdapter(logger interfaces.ILogger) *ManagerAdapter {
	store := utils.NewStore()
	store.StartCleanupGoroutine([]string{"metadata", "action", "signal"}, time.Minute*1)
	return &ManagerAdapter{
		logger:                   logger,
		store:                    store,
		delay:                    time.Second * 1,
		pendingTransactionEvents: make([]dto.TransactionStatus, 0),
		LockedEvents:             false,
	}
}

// Prepare configure manager adapter
func (l *ManagerAdapter) Prepare() error {
	interval, err := utils.GetCollectInterval()
	if err != nil {
		return err
	}
	ymlClient := pkg.NewYmlClient()
	chanLinuxMetrics := make(chan []byte, 1)
	chanClose := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}
	hostStats := pkg.NewHostStats()
	l.interval = interval
	l.ymlClient = ymlClient
	l.hostStats = hostStats
	l.chanMetadata = chanLinuxMetrics
	l.chanClose = chanClose
	l.isClosed = false
	l.wg = wg
	execProgram := pkg.NewExecProgram()
	l.program = execProgram
	osOperation, err := components.SystemOperation(l.logger)
	if err != nil {
		return err
	}
	if err := osOperation.Setup(); err != nil {
		return err
	}
	l.osOperation = osOperation
	fileSystem := pkg.NewFileSystem()
	l.fileSystem = fileSystem
	apiPort, err := utils.GetPortAgentApi()
	if err != nil {
		return err
	}
	l.docpApiPort = apiPort
	agentWorkDir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}
	l.agentWorkDir = agentWorkDir
	stateCheck := services.NewStateCheckService(l.logger)
	if err := stateCheck.Setup(); err != nil {
		return err
	}
	l.stateCheck = stateCheck
	authService := services.NewAuthService(l.logger)
	if err := authService.Setup(); err != nil {
		return err
	}
	l.auth = authService
	utilityService := services.NewUtilityService(l.logger)
	if err := utilityService.Setup(); err != nil {
		return err
	}
	l.utilityService = utilityService
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	l.client = client
	return nil
}

// marshallerMetadata execute marshall the metadata
func (l *ManagerAdapter) marshallerMetadata(metadata dto.Metadata) []byte {
	l.logger.Debug("marshal metadata", "trace", "docp-agent-os-instance.manager_adapter.marshaller")
	bMetadata, err := json.Marshal(metadata)
	if err != nil {
		l.logger.Error("error in marshal metadata", "trace", "docp-agent-os-instance.manager_adapter.marshaller", "error", err.Error())
	}
	return bMetadata
}

// marshaller execute marshal the struct for slice the bytes
func (l *ManagerAdapter) marshaller(inner any) ([]byte, error) {
	l.logger.Debug("execute marshaller", "trace", "docp-agent-os-instance.manager_adapter.marshaller")
	resBytes, err := json.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return resBytes, nil
}

// unmarshaller execute unmarshal the content bytes
func (l *ManagerAdapter) unmarshaller(content []byte, inner any) error {
	l.logger.Debug("execute unmarshaller", "trace", "docp-agent-os-instance.manager_adapter.unmarshaller")
	if err := json.Unmarshal(content, inner); err != nil {
		return err
	}
	return nil
}

// getInfos execute get the infos in host
func (l *ManagerAdapter) getInfos() {
	l.logger.Debug("get infos", "trace", "docp-agent-os-instance.manager_adapter.getInfos")
	computeInfo, err := l.hostStats.ComputeInfo()
	if err != nil {
		l.logger.Error("error host info", "trace", "docp-agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	cpuInfo, err := l.hostStats.CPUInfo()
	if err != nil {
		l.logger.Error("error cpu info", "trace", "docp-agent-os-instance.linux_adapter.getInfos", "error", err.Error())
	}
	memoryInfo, err := l.hostStats.MemoryInfo()
	if err != nil {
		l.logger.Error("error memory info", "trace", "docp-agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	diskInfo, err := l.hostStats.DiskInfo()
	if err != nil {
		l.logger.Error("error disk info", "trace", "docp-agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	processInfo, err := l.hostStats.ProcessInfo()
	if err != nil {
		l.logger.Error("error process info", "trace", "docp-agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	linuxMetadata := dto.Metadata{
		ComputeInfo:  computeInfo,
		CPUInfo:      cpuInfo,
		MemoryInfo:   memoryInfo,
		DiskInfo:     diskInfo,
		ProcessInfos: processInfo,
	}
	if !l.isClosed {
		l.chanMetadata <- l.marshallerMetadata(linuxMetadata)
	}
	l.wg.Done()
}

// firstGetInfos execute first get infos in host
func (l *ManagerAdapter) firstGetInfos() {
	l.logger.Debug("first get infos", "trace", "docp-agent-os-instance.manager_adapter.firstGetInfos")
	l.wg.Add(1)
	time.Sleep(time.Second * 5)
	go l.getInfos()
	l.wg.Done()
}

// closeChannels closing channels
func (l *ManagerAdapter) closeChannels() {
	l.logger.Debug("close channels", "trace", "docp-agent-os-instance.manager_adapter.closeChannels")
	close(l.chanClose)
	close(l.chanMetadata)
	l.isClosed = true
	l.wg.Done()
}

// start execute loop for adapter
func (l *ManagerAdapter) start() {
	l.logger.Debug("start loop", "trace", "docp-agent-os-instance.manager_adapter.start")
	tick := time.NewTicker(l.interval)
	for {
		select {
		case <-l.chanClose:
			break
		case <-tick.C:
			l.wg.Add(1)
			go l.getInfos()
		}
	}
}

// IsLocked return if operator locked for send transactions events
func (l *ManagerAdapter) IsLockedEvents() bool {
	return l.LockedEvents
}

// ExecuteAuthCall execute call to auth and save access token received
func (l *ManagerAdapter) ExecuteAuthCall() error {
	l.logger.Debug("execute auth call", "timestamp", time.Now())
	configAgent, err := l.GetConfigAgent()
	if err != nil {
		return err
	}

	authPayload := dto.AuthPayload{}
	authPayload.ApiKey = configAgent.Agent.ApiKey
	authPayload.ComputeId = configAgent.ComputeId
	resp, statusCode, err := l.auth.AuthCall(authPayload)
	if err != nil {
		return err
	}

	l.logger.Debug("execute auth call", "statusCode", statusCode, "resp", string(resp))
	var authResponse dto.AuthResponse
	switch statusCode {
	case 200:
		if err := l.unmarshaller(resp, &authResponse); err != nil {
			return err
		}

		configAgent.AccessToken = authResponse.AccessToken
		if err := l.UpdateConfigAgent(configAgent); err != nil {
			return err
		}
	}
	return nil
}

// NotifyStatus execute notify the status to state check
func (l *ManagerAdapter) NotifyStatus(status string, typeEvent string, message string, ctx context.Context) error {
	content, err := l.fileSystem.GetFileContent(filepath.Join(l.agentWorkDir, "config.yml"))
	if err != nil {
		return err
	}
	var config dto.ConfigAgent

	if err := l.ymlClient.Unmarshall(content, &config); err != nil {
		return err
	}
	accessToken := config.AccessToken
	if len(accessToken) > 0 {
		transactionStatus := utils.GetTransactionFromContext(ctx)
		if len(transactionStatus.ID) > 0 {
			transactionStatus.Status = status
			transactionStatus.Message = message
			transactionStatus.TypeEvent = typeEvent
			transactionStatus.UlidEvent = utils.GetUlid()
			if l.LockedEvents {
				l.pendingTransactionEvents = append(l.pendingTransactionEvents, transactionStatus)
			} else {
				l.logger.Debug("notify status send status", "transactionStatus", transactionStatus)
				res, statusCode, err := l.stateCheck.SendStatus(transactionStatus)
				if err != nil {
					l.logger.Error("notify status send status", "error", err.Error())
					return err
				}

				switch statusCode {
				case 403:
					if err := l.ExecuteAuthCall(); err != nil {
						l.logger.Error("execute auth call after notify received not authorized", "error", err.Error())
						return err
					}
					// retry last transaction
					res, statusCodeRetry, err := l.stateCheck.SendStatus(transactionStatus)
					if err != nil {
						l.logger.Error("notify status send status", "error", err.Error())
						return err
					}

					l.logger.Debug("retry notify status", "timestamp", time.Now(), "response", string(res), "statusCodeRetry", statusCodeRetry)
				case 500:
					l.LockedEvents = true
					l.pendingTransactionEvents = append(l.pendingTransactionEvents, transactionStatus)
				}

				l.logger.Debug("notify status", "timestamp", time.Now(), "response", string(res), "statusCode", statusCode)

			}
		}
	}

	return nil
}

// Collect execute collect the metrics the host
func (l *ManagerAdapter) Collect() <-chan []byte {
	l.logger.Debug("collect metadata", "trace", "docp-agent-os-instance.manager_adapter.Collect")
	l.wg.Add(2)
	go l.firstGetInfos()
	go l.start()
	return l.chanMetadata
}

// Close closing collect loop
func (l *ManagerAdapter) Close() error {
	l.logger.Debug("execute close", "trace", "docp-agent-os-instance.manager_adapter.Close")
	l.chanClose <- struct{}{}
	l.wg.Add(1)
	go l.closeChannels()
	return nil
}

// GetAgentVersion return version installed agent
func (l *ManagerAdapter) GetAgentVersion() (string, error) {
	var configAgent dto.ConfigAgent
	configPath, err := utils.GetConfigFilePath()
	if err != nil {
		return "", err
	}

	content, err := l.fileSystem.GetFileContent(configPath)
	if err != nil {
		return "", err
	}

	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return "", err
	}

	return configAgent.Version, nil
}

// GetAgentRollbackVersion return rollback version installed agent
func (l *ManagerAdapter) GetAgentRollbackVersion() (string, error) {
	var configAgent dto.ConfigAgent
	configPath, err := utils.GetConfigFilePath()
	if err != nil {
		return "", err
	}

	content, err := l.fileSystem.GetFileContent(configPath)
	if err != nil {
		return "", err
	}

	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return "", err
	}

	return configAgent.RollbackVersion, nil
}

// GetAgentVersionFromSignalBytes return version from signal bytes
func (l *ManagerAdapter) GetAgentVersionFromSignalBytes(data []byte) (string, error) {
	var stateCheckResponse dto.StateCheckResponse
	var version string
	if err := l.unmarshaller(data, &stateCheckResponse); err != nil {
		return version, err
	}

	if len(stateCheckResponse.Signal.Agents.DocpAgent.Version) > 0 {
		version = stateCheckResponse.Signal.Agents.DocpAgent.Version
	}
	return version, nil
}

// SaveAgentVersion save version installed agent
func (l *ManagerAdapter) SaveAgentVersion(version string) error {
	var configAgent dto.ConfigAgent
	configPath, err := utils.GetConfigFilePath()
	if err != nil {
		return err
	}

	content, err := l.fileSystem.GetFileContent(configPath)
	if err != nil {
		return err
	}

	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return err
	}

	configAgent.Version = version

	ymlBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(configPath, ymlBytes); err != nil {
		return err
	}

	return nil
}

// SaveAgentRollbackVersion save rollback version installed agent
func (l *ManagerAdapter) SaveAgentRollbackVersion(version string) error {
	var configAgent dto.ConfigAgent
	configPath, err := utils.GetConfigFilePath()
	if err != nil {
		return err
	}

	content, err := l.fileSystem.GetFileContent(configPath)
	if err != nil {
		return err
	}

	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return err
	}

	configAgent.RollbackVersion = version

	ymlBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(configPath, ymlBytes); err != nil {
		return err
	}

	return nil
}

// Status return status from systemd api
func (l *ManagerAdapter) Status(serviceName string) (string, error) {
	l.logger.Debug("execute verify status the service", "trace", "docp-agent-os-instance.manager_adapter.Status", "serviceName", serviceName)
	output, err := l.osOperation.Status(serviceName)
	if err != nil {
		return "", err
	}
	output = strings.ReplaceAll(output, "\"", "")
	l.logger.Debug("execute verify status the service", "trace", "docp-agent-os-instance.manager_adapter.Status", "serviceName", serviceName, "output", output)
	return output, nil
}

// AlreadyInstalled return if service already installed
func (l *ManagerAdapter) AlreadyInstalled(serviceName string) (bool, error) {
	l.logger.Debug("execute verify already installed service", "trace", "docp-agent-os-instance.manager_adapter.AlreadyInstalled", "serviceName", serviceName)
	installed, err := l.osOperation.AlreadyInstalled(serviceName)
	if err != nil {
		return false, err
	}
	l.logger.Debug("execute verify already installed service", "trace", "docp-agent-os-instance.manager_adapter.AlreadyInstalled", "serviceName", serviceName, "installed", installed)
	return installed, nil
}

// DaemonReload execute daemon reload the service in systemd
func (l *ManagerAdapter) DaemonReload() error {
	l.logger.Debug("daemon reload", "trace", "docp-agent-os-instance.manager_adapter.DaemonReload")
	if err := l.osOperation.DaemonReload(); err != nil {
		return err
	}
	return nil
}

// RestartService execute restart the service in systemd
func (l *ManagerAdapter) RestartService(serviceName string) error {
	l.logger.Debug("restart service", "trace", "docp-agent-os-instance.manager_adapter.RestartService")
	if err := l.osOperation.RestartService(serviceName); err != nil {
		return err
	}
	return nil
}

// StopService execute stop the service in systemd
func (l *ManagerAdapter) StopService(serviceName string) error {
	l.logger.Debug("stop service", "trace", "docp-agent-os-instance.manager_adapter.StopService")
	if err := l.osOperation.StopService(serviceName); err != nil {
		return err
	}
	return nil
}

// InstallAgent execute install the agent docp
func (l *ManagerAdapter) InstallAgent(version string) error {
	l.logger.Debug("install agent", "trace", "docp-agent-os-instance.manager_adapter.InstallAgent")
	if err := l.osOperation.InstallAgent(version); err != nil {
		return err
	}
	return nil
}

// UninstallAgent execute uninstall the agent docp
func (l *ManagerAdapter) UninstallAgent() error {
	l.logger.Debug("uninstall agent", "trace", "docp-agent-os-instance.manager_adapter.UninstallAgent")
	if err := l.osOperation.UninstallAgent(); err != nil {
		return err
	}
	return nil
}

// UpdateAgent execute update the agent docp
func (l *ManagerAdapter) UpdateAgent(version string) error {
	l.logger.Debug("update agent", "trace", "docp-agent-os-instance.manager_adapter.UpdateAgent")
	if version == "latest" {
		agentVersions, err := l.FetchAgentVersions()
		if err != nil {
			return err
		}
		version = agentVersions.LatestVersion
	}
	if err := l.osOperation.UpdateAgent(version); err != nil {
		return err
	}
	return nil
}

// AutoUninstall execute auto uninstall the manager
func (l *ManagerAdapter) AutoUninstall() error {
	l.logger.Debug("auto uninstall agent", "trace", "docp-agent-os-instance.manager_adapter.AutoUninstall")
	if err := l.osOperation.AutoUninstall(); err != nil {
		return err
	}
	return nil
}

// FetchAgentVersions fetches the available agent versions
func (l *ManagerAdapter) FetchAgentVersions() (dto.AgentVersions, error) {
	l.logger.Debug("fetch agent versions", "trace", "docp-agent-os-instance.manager_adapter.FetchAgentVersions")
	agentVersions, err := l.utilityService.FetchAgentVersions()
	if err != nil {
		return dto.AgentVersions{}, err
	}
	return agentVersions, nil
}

// requestForAgentInstallDatadog execute request for agent
func (l *ManagerAdapter) requestForAgentInstallDatadog(url string, method string, data []byte) ([]byte, error) {
	l.logger.Debug("request for agent install datadog", "trace", "docp-agent-os-instance.manager_adapter.requestForAgentInstallDatadog", "url", url, "method", method, "data", data)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		l.logger.Error("error in create request", "trace", "docp-agent-os-instance.manager_adapter.requestForAgentInstallDatadog", "error", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := l.client.Do(req)
	if err != nil {
		l.logger.Error("error in execute request", "trace", "docp-agent-os-instance.manager_adapter.requestForAgentInstallDatadog", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		l.logger.Error("error in read body response", "trace", "docp-agent-os-instance.manager_adapter.requestForAgentInstallDatadog", "error", err.Error())
		return nil, err
	}
	return respBytes, nil
}

// DocpAgentApiInstallDatadog execute call to api docp for install datadog agent
func (l *ManagerAdapter) DocpAgentApiInstallDatadog(ddApiKey, ddSite string) ([]byte, error) {
	l.logger.Debug("execute send request for install datadog agent", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiInstallDatadog", "ddApiKey", ddApiKey, "ddSite", ddSite)

	transaction := utils.NewTransactionStatus()
	ctx := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("install_docp_vendor_received", pkg.TransactionEventOpen, "install docp vendor received", ctx)
	time.Sleep(l.delay)

	datadogDto := dto.DatadogInstallDTO{
		DDSite:   ddSite,
		DDApiKey: ddApiKey,
	}
	bDatadogDto, err := l.marshaller(&datadogDto)
	if err != nil {
		go l.NotifyStatus("install_docp_vendor_error", pkg.TransactionEventClose, "failed install docp vendor", ctx)
		return nil, err
	}
	go l.NotifyStatus("install_docp_vendor_processing", pkg.TransactionEventUpdate, "install docp vendor processing", ctx)
	time.Sleep(l.delay)

	urlDocpInstallDatadog := fmt.Sprintf("http://127.0.0.1:%s/datadog/install", l.docpApiPort)
	respBytes, err := l.requestForAgentInstallDatadog(urlDocpInstallDatadog, http.MethodPost, bDatadogDto)
	if err != nil {
		go l.NotifyStatus("install_docp_vendor_error", pkg.TransactionEventClose, "failed install docp vendor", ctx)
		l.logger.Error("error in read body response", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiInstallDatadog", "error", err.Error())
		return nil, err
	}

	go l.NotifyStatus("install_docp_vendor_completed", pkg.TransactionEventClose, "install docp vendor completed", ctx)
	return respBytes, nil
}

// DocpAgentApiInstallDatadog execute call to api docp for install datadog agent
func (l *ManagerAdapter) DocpAgentApiInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries string) ([]byte, error) {
	l.logger.Debug("execute send request for install datadog agent with apm single step", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiInstallDatadogWithApmSingleStep", "ddApiKey", ddApiKey, "ddSite", ddSite, "ddEnv", ddEnv, "ddApmInstrumentationEnabled", ddApmInstrumentationEnabled, "ddApmInstrumentationLibraries", ddApmInstrumentationLibraries)

	transaction := utils.NewTransactionStatus()
	ctx := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("install_docp_vendor_tracer_received", pkg.TransactionEventOpen, "install docp vendor tracer single step received", ctx)
	time.Sleep(l.delay)

	datadogDto := dto.DatadogInstallDTO{
		DDSite:    ddSite,
		DDApiKey:  ddApiKey,
		Component: "tracer",
		Mode:      "single_step",
		EnvVars: []dto.DatadogEnvVars{
			{
				Name:  "DD_APM_INSTRUMENTATION_ENABLED",
				Value: ddApmInstrumentationEnabled,
			},
			{
				Name:  "DD_APM_INSTRUMENTATION_LIBRARIES",
				Value: ddApmInstrumentationLibraries,
			},
			{
				Name:  "DD_ENV",
				Value: ddEnv,
			},
		},
	}
	bDatadogDto, err := l.marshaller(&datadogDto)
	if err != nil {
		go l.NotifyStatus("install_docp_vendor_tracer_error", pkg.TransactionEventClose, "failed install docp vendor tracer single step", ctx)
		return nil, err
	}

	urlDocpInstallDatadog := fmt.Sprintf("http://127.0.0.1:%s/datadog/install", l.docpApiPort)

	go l.NotifyStatus("install_docp_vendor_tracer_processing", pkg.TransactionEventUpdate, "install docp vendor tracer single step processing", ctx)
	time.Sleep(l.delay)

	respBytes, err := l.requestForAgentInstallDatadog(urlDocpInstallDatadog, http.MethodPost, bDatadogDto)
	if err != nil {
		go l.NotifyStatus("install_docp_vendor_tracer_error", pkg.TransactionEventClose, "failed install docp vendor tracer single step", ctx)
		l.logger.Error("error in read body response", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiInstallDatadogWithApmSingleStep", "error", err.Error())
		return nil, err
	}

	if err := l.SaveAlreadyTracer(true); err != nil {
		l.logger.Error("error in read body response", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiInstallDatadogWithApmSingleStep", "error", err.Error())
		go l.NotifyStatus("install_docp_vendor_tracer_error", pkg.TransactionEventClose, "failed install docp vendor tracer single step", ctx)
		return nil, err
	}

	go l.NotifyStatus("install_docp_vendor_tracer_completed", pkg.TransactionEventClose, "install docp vendor tracer single step completed", ctx)
	return respBytes, nil
}

// DocpAgentApiInstallDatadog execute call to api docp for install datadog agent
func (l *ManagerAdapter) DocpAgentApiInstallDatadogWithApmTracingLibrary(ddApiKey, ddSite, language, pathTracer, version string) ([]byte, error) {
	l.logger.Debug("execute send request for install datadog agent with apm single step", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiInstallDatadogWithApmSingleStep", "ddApiKey", ddApiKey, "ddSite", ddSite, "language", language, "pathTracer", pathTracer, "version", version)

	transaction := utils.NewTransactionStatus()
	ctx := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("install_docp_vendor_tracer_received", pkg.TransactionEventOpen, "install docp vendor tracer library received", ctx)
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
	bDatadogDto, err := l.marshaller(&datadogDto)
	if err != nil {
		go l.NotifyStatus("install_docp_vendor_tracer_error", pkg.TransactionEventClose, "failed install docp vendor tracer", ctx)
		return nil, err
	}
	urlDocpInstallDatadog := fmt.Sprintf("http://127.0.0.1:%s/datadog/tracer/install", l.docpApiPort)

	go l.NotifyStatus("install_docp_vendor_tracer_processing", pkg.TransactionEventUpdate, "install docp vendor tracer library processing", ctx)
	time.Sleep(l.delay)

	respBytes, err := l.requestForAgentInstallDatadog(urlDocpInstallDatadog, http.MethodPost, bDatadogDto)
	if err != nil {
		l.logger.Error("error in read body response", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiInstallDatadogWithApmSingleStep", "error", err.Error())
		go l.NotifyStatus("install_docp_vendor_tracer_error", pkg.TransactionEventClose, "failed install docp vendor tracer", ctx)
		return nil, err
	}
	if err := l.SaveAlreadyTracer(true); err != nil {
		l.logger.Error("error in read body response", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiInstallDatadogWithApmSingleStep", "error", err.Error())
		go l.NotifyStatus("install_docp_vendor_tracer_error", pkg.TransactionEventClose, "failed install docp vendor tracer", ctx)
		return nil, err
	}

	go l.NotifyStatus("install_docp_vendor_tracer_completed", pkg.TransactionEventClose, "install docp vendor tracer library completed", ctx)
	return respBytes, nil
}

// DocpAgentApiUninstallDatadog execute call to api docp for uninstall datadog agent
func (l *ManagerAdapter) DocpAgentApiUninstallDatadog() ([]byte, error) {
	l.logger.Debug("execute send request for uninstall datadog agent", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUninstallDatadog")

	transaction := utils.NewTransactionStatus()
	ctxTransaction := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("uninstall_docp_vendor_received", pkg.TransactionEventOpen, "uninstall docp vendor received", ctxTransaction)
	time.Sleep(l.delay)

	urlDocpUninstallDatadog := fmt.Sprintf("http://127.0.0.1:%s/datadog/uninstall", l.docpApiPort)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlDocpUninstallDatadog, nil)
	if err != nil {
		go l.NotifyStatus("uninstall_docp_vendor_error", pkg.TransactionEventClose, "failed uninstall vendor", ctxTransaction)
		l.logger.Error("error in create request", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUninstallDatadog", "error", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := l.client.Do(req)
	if err != nil {
		go l.NotifyStatus("uninstall_docp_vendor_error", pkg.TransactionEventClose, "failed uninstall vendor", ctxTransaction)
		l.logger.Error("error in execute request", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUninstallDatadog", "error", err.Error())
		return nil, err
	}

	go l.NotifyStatus("uninstall_docp_vendor_processing", pkg.TransactionEventUpdate, "uninstall docp vendor processing", ctxTransaction)
	time.Sleep(l.delay)

	go l.NotifyStatus("uninstall_docp_vendor_complete", pkg.TransactionEventClose, "uninstall docp vendor completed", ctxTransaction)
	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		go l.NotifyStatus("uninstall_docp_vendor_error", pkg.TransactionEventClose, "failed uninstall vendor", ctxTransaction)
		l.logger.Error("error in read body response", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUninstallDatadog", "error", err.Error())
		return nil, err
	}
	if err := l.SaveAlreadyTracer(false); err != nil {
		go l.NotifyStatus("uninstall_docp_vendor_error", pkg.TransactionEventClose, "failed uninstall vendor", ctxTransaction)
		l.logger.Error("error in save already tracer", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUninstallDatadog", "error", err.Error())
		return nil, err
	}

	return respBytes, nil
}

// DocpAgentApiUpdateConfigurationsDatadog execute call to api docp for update datadog configurations
func (l *ManagerAdapter) DocpAgentApiUpdateConfigurationsDatadog(content []byte) ([]byte, error) {
	l.logger.Debug("execute send request for update configurations in datadog agent", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUpdateConfigurationsDatadog", "content", string(content))

	transaction := utils.NewTransactionStatus()
	ctxTransaction := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.NotifyStatus("update_vendor_received", pkg.TransactionEventOpen, "update vendor received", ctxTransaction)
	time.Sleep(l.delay)

	urlDocpUpdateConfigurations := fmt.Sprintf("http://127.0.0.1:%s/datadog/configurations", l.docpApiPort)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlDocpUpdateConfigurations, bytes.NewBuffer(content))
	if err != nil {
		go l.NotifyStatus("update_docp_vendor_error", pkg.TransactionEventClose, "failed update datadog", ctxTransaction)
		l.logger.Error("error in create request", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUpdateConfigurationsDatadog", "error", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := l.client.Do(req)
	if err != nil {
		go l.NotifyStatus("update_docp_vendor_error", pkg.TransactionEventClose, "failed update datadog", ctxTransaction)
		l.logger.Error("error in execute request", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUpdateConfigurationsDatadog", "error", err.Error())
		return nil, err
	}

	go l.NotifyStatus("update_docp_vendor_processing", pkg.TransactionEventUpdate, "update docp vendor processing", ctxTransaction)
	time.Sleep(l.delay)

	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		go l.NotifyStatus("update_docp_vendor_error", pkg.TransactionEventClose, "failed update datadog", ctxTransaction)
		l.logger.Error("error in read body response", "trace", "docp-agent-os-instance.manager_adapter.DocpAgentApiUpdateConfigurationsDatadog", "error", err.Error())
		return nil, err
	}

	go l.NotifyStatus("update_docp_vendor_complete", pkg.TransactionEventClose, "update docp vendor completed", ctxTransaction)
	return respBytes, nil
}

// GetStateReceived return state received from state check service
func (l *ManagerAdapter) GetStateReceived() ([]byte, error) {
	l.logger.Debug("get state received", "trace", "docp-agent-os-instance.manager_adapter.GetStateReceived")
	res, err := l.fileSystem.GetFileContent(filepath.Join(l.agentWorkDir, "state", "received"))
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetStateCurrent return state current from file
func (l *ManagerAdapter) GetStateCurrent() ([]byte, error) {
	l.logger.Debug("get state current", "trace", "docp-agent-os-instance.manager_adapter.GetStateCurrent")
	res, err := l.fileSystem.GetFileContent(filepath.Join(l.agentWorkDir, "state", "current"))
	if err != nil {
		return nil, err
	}
	return res, nil
}

// SaveStateReceived execute save the state received
func (l *ManagerAdapter) SaveStateReceived(stateData []byte) error {
	l.logger.Debug("save state received", "trace", "docp-agent-os-instance.manager_adapter.SaveStateReceived", "stateData", string(stateData))
	if err := l.fileSystem.WriteFileContent(filepath.Join(l.agentWorkDir, "state", "received"), stateData); err != nil {
		return err
	}
	return nil
}

// SaveStateCurrent execute save the state current
func (l *ManagerAdapter) SaveStateCurrent(stateData []byte) error {
	l.logger.Debug("save state current", "trace", "docp-agent-os-instance.manager_adapter.SaveStateCurrent", "stateData", string(stateData))
	if err := l.fileSystem.WriteFileContent(filepath.Join(l.agentWorkDir, "state", "current"), stateData); err != nil {
		return err
	}
	return nil
}

// ExisteOtherVendors verify if exist other vendors
func (l *ManagerAdapter) ExisteOtherVendors() (bool, error) {
	vendors, err := l.GetRemoveOtherVendors()
	if err != nil {
		return false, err
	}
	if vendors != nil && len(vendors) > 0 {
		return true, nil
	}
	return false, nil
}

// GetRemoveOtherVendors return other vendors from signal
func (l *ManagerAdapter) GetRemoveOtherVendors() ([]string, error) {
	var state dto.StateCheckResponse
	received, err := l.GetStateReceived()
	if err != nil {
		return nil, err
	}

	if err := l.unmarshaller(received, &state); err != nil {
		return nil, err
	}

	otherVendors := state.Signal.RemoveOtherVendors

	return otherVendors, nil
}

// prepareDocpAgentAction execute prepare for docp agent action
func (l *ManagerAdapter) prepareDocpAgentAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare docp agent action", "trace", "docp-agent-os-instance.manager_adapter.prepareDocpAgentAction", "stateCheckSignal", stateCheckSignal)
	if stateCheckSignal.TypeSignal == "update" {
		if len(stateCheckSignal.Agents.DocpAgent.Version) > 0 {
			return dto.StateAction{
				Type:    "docp-agent",
				Action:  "update",
				Version: stateCheckSignal.Agents.DocpAgent.Version,
			}
		}
	} else if stateCheckSignal.TypeSignal == "uninstall" {
		return dto.StateAction{
			Type:   "docp-agent",
			Action: "uninstall",
		}
	}
	return dto.StateAction{}
}

// getActionsEnvs return slice the envs from configurations
func (l *ManagerAdapter) getActionsEnvs(envs []dto.StateCheckEnvVars) []dto.StateActionEnvs {
	l.logger.Debug("get actions envs", "trace", "docp-agent-os-instance.manager_adapter.getActionsEnvs", "envs", envs)
	var arrEnvs []dto.StateActionEnvs
	if len(envs) > 0 {
		for _, ev := range envs {
			enAct := dto.StateActionEnvs{
				Name:  ev.Name,
				Value: ev.Value,
			}
			arrEnvs = append(arrEnvs, enAct)
		}
	}
	return arrEnvs
}

// getActionsFiles return slice the files from configurations
func (l *ManagerAdapter) getActionsFiles(configurations dto.StateCheckDatadogConfigurations) []dto.StateActionFiles {
	l.logger.Debug("get actions files", "trace", "docp-agent-os-instance.manager_adapter.getActionsFiles", "configurations", configurations)
	var arrFiles []dto.StateActionFiles
	if len(configurations.Files) > 0 {
		for _, fl := range configurations.Files {
			flAct := dto.StateActionFiles{
				FilePath: fl.FilePath,
				Content:  fl.Content,
			}
			arrFiles = append(arrFiles, flAct)
		}
	}
	return arrFiles
}

func (l *ManagerAdapter) extractComponentEnvVarsApm(installWithEnvVars []dto.StateCheckEnvVars) []dto.StateActionEnvs {
	l.logger.Debug("extract component env vars apm", "trace", "docp-agent-os-instance.manager_adapter.extractComponentEnvVarsApm", "installWithEnvVars", installWithEnvVars)
	actionsEnvs := []dto.StateActionEnvs{}
	for _, env := range installWithEnvVars {
		newEnv := dto.StateActionEnvs{
			Name:  env.Name,
			Value: env.Value,
		}
		actionsEnvs = append(actionsEnvs, newEnv)
	}
	return actionsEnvs
}

// parseSiteDatadog execute parse the site datadog
func (l *ManagerAdapter) parseSiteDatadog(site string) string {
	l.logger.Debug("parse site datadog", "trace", "docp-agent-os-instance.manager_adapter.parseSiteDatadog", "site", site)
	host, err := utils.GetBaseUrlSite(site)
	if err != nil {
		return "datadoghq.com"
	}
	return host
}

// prepareTracerDatadogSingleStepAction execute prepare for tracer datadog single step action
func (l *ManagerAdapter) prepareTracerDatadogSingleStepAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare tracer datadog library action", "trace", "docp-agent-os-instance.manager_adapter.prepareTracerDatadogSingleStepAction", "stateCheckSignal", stateCheckSignal)
	var action dto.StateAction
	var envVars []dto.StateActionEnvs
	if stateCheckSignal.TypeSignal == "update" {
		if len(stateCheckSignal.Agents.DatadogTracerSingleStep.Version) > 0 && len(stateCheckSignal.Agents.DatadogTracerLibrary.Version) == 0 {
			tracerSingleStep := stateCheckSignal.Agents.DatadogTracerSingleStep
			var files []dto.StateActionFiles
			var componetEnvVars []dto.StateActionEnvs

			componetEnvVars = l.extractComponentEnvVarsApm(tracerSingleStep.InstallWithEnvVars)

			if len(stateCheckSignal.Agents.DatadogAgent.Version) > 0 {
				datadogAgent := stateCheckSignal.Agents.DatadogAgent

				envVars = append(envVars, dto.StateActionEnvs{
					Name:  "DD_API_KEY",
					Value: datadogAgent.ApiKey,
				})
				envVars = append(envVars, dto.StateActionEnvs{
					Name:  "DD_APP_KEY",
					Value: datadogAgent.AppKey,
				})
				envVars = append(envVars, dto.StateActionEnvs{
					Name:  "DD_SITE",
					Value: l.parseSiteDatadog(datadogAgent.Site),
				})
				files = l.getActionsFiles(datadogAgent.Configurations)
			}

			action = dto.StateAction{
				Type:          "datadog",
				Action:        "install",
				Mode:          "single_step",
				Component:     "tracer",
				ComponentEnvs: componetEnvVars,
				Envs:          envVars,
				Files:         files,
			}
		}
	}
	return action
}

// prepareTracerDatadogLibraryAction execute prepare for tracer datadog tracer library action
func (l *ManagerAdapter) prepareTracerDatadogLibraryAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare tracer datadog library action", "trace", "docp-agent-os-instance.manager_adapter.prepareTracerDatadogLibraryAction", "stateCheckSignal", stateCheckSignal)
	var action dto.StateAction
	var envVars []dto.StateActionEnvs
	if stateCheckSignal.TypeSignal == "update" {
		if len(stateCheckSignal.Agents.DatadogTracerLibrary.Version) > 0 && len(stateCheckSignal.Agents.DatadogTracerSingleStep.Version) == 0 {
			tracerLibrary := stateCheckSignal.Agents.DatadogTracerLibrary
			var files []dto.StateActionFiles
			var componetEnvVars []dto.StateActionEnvs

			componetEnvVars = append(componetEnvVars, dto.StateActionEnvs{
				Name:  "language",
				Value: tracerLibrary.Language,
			})
			componetEnvVars = append(componetEnvVars, dto.StateActionEnvs{
				Name:  "path_tracer",
				Value: tracerLibrary.PathTracer,
			})
			componetEnvVars = append(componetEnvVars, dto.StateActionEnvs{
				Name:  "version",
				Value: tracerLibrary.Version,
			})

			if len(stateCheckSignal.Agents.DatadogAgent.Version) > 0 {
				datadogAgent := stateCheckSignal.Agents.DatadogAgent

				envVars = append(envVars, dto.StateActionEnvs{
					Name:  "DD_API_KEY",
					Value: datadogAgent.ApiKey,
				})
				envVars = append(envVars, dto.StateActionEnvs{
					Name:  "DD_APP_KEY",
					Value: datadogAgent.AppKey,
				})
				envVars = append(envVars, dto.StateActionEnvs{
					Name:  "DD_SITE",
					Value: l.parseSiteDatadog(datadogAgent.Site),
				})
				files = l.getActionsFiles(datadogAgent.Configurations)
			}

			action = dto.StateAction{
				Type:          "datadog",
				Action:        "install",
				Mode:          "tracing_library",
				Component:     "tracer",
				ComponentEnvs: componetEnvVars,
				Envs:          envVars,
				Files:         files,
			}
		}
	}
	return action
}

// prepareAgentDatadogAction execute prepare for agent datadog action
func (l *ManagerAdapter) prepareAgentDatadogAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare agent datadog action", "trace", "docp-agent-os-instance.manager_adapter.prepareAgentDatadogAction", "stateCheckSignal", stateCheckSignal)
	var action dto.StateAction
	var envVars []dto.StateActionEnvs
	var componetEnvVars []dto.StateActionEnvs
	var files []dto.StateActionFiles

	if stateCheckSignal.TypeSignal == "update" {
		if len(stateCheckSignal.Agents.DatadogAgent.Version) > 0 {
			datadogAgent := stateCheckSignal.Agents.DatadogAgent

			envVars = append(envVars, dto.StateActionEnvs{
				Name:  "DD_API_KEY",
				Value: datadogAgent.ApiKey,
			})
			envVars = append(envVars, dto.StateActionEnvs{
				Name:  "DD_APP_KEY",
				Value: datadogAgent.AppKey,
			})
			envVars = append(envVars, dto.StateActionEnvs{
				Name:  "DD_SITE",
				Value: l.parseSiteDatadog(datadogAgent.Site),
			})
			files = l.getActionsFiles(datadogAgent.Configurations)
			l.logger.Debug("prepare agent datadog action", "files", files)
			return dto.StateAction{
				Type:          "datadog",
				Action:        "install",
				Mode:          "",
				Component:     "agent",
				ComponentEnvs: componetEnvVars,
				Envs:          envVars,
				Files:         files,
			}
		}
	} else if stateCheckSignal.TypeSignal == "uninstall" {
		if len(stateCheckSignal.RemoveOtherVendors) > 0 {
			for _, vendor := range stateCheckSignal.RemoveOtherVendors {
				if vendor == "datadog" {
					return dto.StateAction{
						Type:      "datadog",
						Action:    "uninstall",
						Mode:      "",
						Component: "agent",
					}
				}
			}
		}
	}
	return action
}

// prepareAgentDatadogUpdateAction execute prepare for agent datadog update action
func (l *ManagerAdapter) prepareAgentDatadogUpdateAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare agent datadog action", "trace", "docp-agent-os-instance.manager_adapter.prepareAgentDatadogAction", "stateCheckSignal", stateCheckSignal)
	var action dto.StateAction
	var envVars []dto.StateActionEnvs
	var componetEnvVars []dto.StateActionEnvs
	var files []dto.StateActionFiles

	if stateCheckSignal.TypeSignal == "update" {
		if len(stateCheckSignal.Agents.DatadogAgent.Version) > 0 {
			datadogAgent := stateCheckSignal.Agents.DatadogAgent

			files = l.getActionsFiles(datadogAgent.Configurations)
			l.logger.Debug("prepare agent datadog update action", "files", files)
			act := dto.StateAction{
				Type:          "datadog",
				Action:        "update",
				Mode:          "",
				Component:     "agent",
				ComponentEnvs: componetEnvVars,
				Envs:          envVars,
				Files:         files,
			}
			action = act
		}
	}
	return action
}

// removeAgentDatadogIfTracerSingleStepExists remove agent if tracer single step exists
func (l *ManagerAdapter) removeAgentDatadogIfTracerSingleStepExists(arrStateActions []dto.StateAction) []dto.StateAction {
	newArrStateActions := []dto.StateAction{}
	agentExists := false
	tracerSingleStepExists := false
	for _, act := range arrStateActions {
		if act.Type == "datadog" {
			if act.Component == "agent" {
				agentExists = true
			} else if act.Component == "tracer" && act.Mode == "single_step" {
				tracerSingleStepExists = true
			}
		}
	}
	if agentExists && tracerSingleStepExists {
		for _, act := range arrStateActions {
			if act.Type == "datadog" && act.Component == "agent" {
				continue
			}
			newArrStateActions = append(newArrStateActions, act)
		}
	} else {
		return arrStateActions
	}
	return newArrStateActions
}

// GetActions return actions for agents
func (l *ManagerAdapter) GetActions(stateCheckResponse *dto.StateCheckResponse) ([]dto.StateAction, error) {
	l.logger.Debug("get actions", "trace", "docp-agent-os-instance.manager_adapter.GetActions", "stateCheckResponse", stateCheckResponse)
	var arrStateActions []dto.StateAction

	// prepare docp agent action
	docpAgentAction := l.prepareDocpAgentAction(stateCheckResponse.Signal)
	lastDocpAgentActionHash := l.GetStore("action.docp.state")
	docpAgentActionBytes, err := l.marshaller(&docpAgentAction)
	if err != nil {
		return nil, err
	}

	newDocpAgentActionHash := utils.GenerateMd5Hash(docpAgentActionBytes)
	if newDocpAgentActionHash != lastDocpAgentActionHash {
		arrStateActions = append(arrStateActions, docpAgentAction)
		if err := l.SetStore("action.docp.state", newDocpAgentActionHash); err != nil {
			return nil, err
		}
	}
	// validate datadog installed
	alreadyInstalled, err := l.AlreadyInstalled("datadog")
	if err != nil {
		return nil, err
	}

	// prepare datadog agent action
	agentDatadogAction := l.prepareAgentDatadogAction(stateCheckResponse.Signal)
	lastAgentDatadogActionHash := l.GetStore("action.datadog.state")
	datadogAgentActionBytes, err := l.marshaller(&agentDatadogAction)
	if err != nil {
		return nil, err
	}

	newAgentDatadogActionHash := fmt.Sprintf("%s.%v", utils.GenerateMd5Hash(datadogAgentActionBytes), alreadyInstalled)
	if newAgentDatadogActionHash != lastAgentDatadogActionHash {
		arrStateActions = append(arrStateActions, agentDatadogAction)
		if err := l.SetStore("action.datadog.state", newAgentDatadogActionHash); err != nil {
			return nil, err
		}
	}

	// prepare datadog update agent action
	agentDatadogUpdateAction := l.prepareAgentDatadogUpdateAction(stateCheckResponse.Signal)
	lastAgentDatadogUpdateActionHash := l.GetStore("action.datadog.update")
	datadogUpdateAgentActionBytes, err := l.marshaller(&agentDatadogUpdateAction)
	if err != nil {
		return nil, err
	}

	newAgentDatadogUpdateActionHash := utils.GenerateMd5Hash(datadogUpdateAgentActionBytes)
	if newAgentDatadogUpdateActionHash != lastAgentDatadogUpdateActionHash {
		if alreadyInstalled {
			arrStateActions = append(arrStateActions, agentDatadogUpdateAction)
			if err := l.SetStore("action.datadog.update", newAgentDatadogUpdateActionHash); err != nil {
				return nil, err
			}
		}
	}

	// prepare datadog tracer library action
	tracerDatadogLibraryAction := l.prepareTracerDatadogLibraryAction(stateCheckResponse.Signal)
	lastTracerDatadogLibraryActionHash := l.GetStore("action.datadog.tracer.library")
	tracerDatadogLibraryActionBytes, err := l.marshaller(&tracerDatadogLibraryAction)
	if err != nil {
		return nil, err
	}

	newTracerDatadogLibraryActionHash := utils.GenerateMd5Hash(tracerDatadogLibraryActionBytes)
	if newTracerDatadogLibraryActionHash != lastTracerDatadogLibraryActionHash {
		arrStateActions = append(arrStateActions, tracerDatadogLibraryAction)
		if err := l.SetStore("action.datadog.tracer.library", newTracerDatadogLibraryActionHash); err != nil {
			return nil, err
		}
	}

	// prepare datadog tracer single step action
	tracerDatadogSingleStepAction := l.prepareTracerDatadogSingleStepAction(stateCheckResponse.Signal)
	lastTracerDatadogSingleStepActionHash := l.GetStore("action.datadog.tracer.single.step")
	tracerDatadogSingleStepActionBytes, err := l.marshaller(&tracerDatadogSingleStepAction)
	if err != nil {
		return nil, err
	}

	newTracerDatadogSingleStepActionHash := utils.GenerateMd5Hash(tracerDatadogSingleStepActionBytes)
	if newTracerDatadogSingleStepActionHash != lastTracerDatadogSingleStepActionHash {
		arrStateActions = append(arrStateActions, tracerDatadogSingleStepAction)
		if err := l.SetStore("action.datadog.tracer.single.step", newTracerDatadogSingleStepActionHash); err != nil {
			return nil, err
		}
	}

	arrStateActionsFiltered := l.removeAgentDatadogIfTracerSingleStepExists(arrStateActions)
	l.logger.Debug("get actions", "trace", "docp-agent-os-instance.manager_adapter.GetActions", "docpAgentAction", docpAgentAction, "agentDatadogAction", agentDatadogAction, "agentDatadogUpdateAction", agentDatadogUpdateAction, "tracerDatadogLibraryAction", tracerDatadogLibraryAction, "tracerDatadogSingleStepAction", tracerDatadogSingleStepAction)
	l.logger.Debug("get actions", "trace", "docp-agent-os-instance.manager_adapter.GetActions", "arrStateActions", arrStateActions)

	return arrStateActionsFiltered, nil
}

// SaveState save state from state check
func (l *ManagerAdapter) SaveState(data []byte) error {
	l.logger.Debug("save state", "trace", "docp-agent-os-instance.manager_adapter.SaveState", "data", string(data))
	if err := l.SaveStateReceived(data); err != nil {
		return err
	}
	return nil
}

// GetState get state from state check
func (l *ManagerAdapter) GetState() ([]byte, error) {
	l.logger.Debug("get state", "trace", "docp-agent-os-instance.manager_adapter.GetState")
	var stateData dto.StateCheckResponse

	// get state received
	pathStateReceived := filepath.Join(l.agentWorkDir, "state", "received")
	content, err := l.fileSystem.GetFileContent(pathStateReceived)
	if err != nil {
		return nil, err
	}

	if err := l.unmarshaller(content, &stateData); err != nil {
		return nil, err
	}

	stateActions, err := l.GetActions(&stateData)
	if err != nil {
		return nil, err
	}

	bStateActions, err := l.marshaller(&stateActions)
	if err != nil {
		return nil, err
	}

	return bStateActions, nil
}

// GetDatadogStateFromReceived get state datadog from received file
func (l *ManagerAdapter) GetDatadogStateFromReceived() (string, error) {
	l.logger.Debug("get datadog state from received", "trace", "docp-agent-os-instance.manager_adapter.GetDatadogStateFromReceived")
	var agentState dto.StateCheckResponse

	content, err := l.GetStateReceived()
	if err != nil {
		return "", err
	}

	if err := l.unmarshaller(content, &agentState); err != nil {
		return "", err
	}

	stateSignal := agentState.Signal

	if len(stateSignal.Agents.DatadogAgent.Version) == 0 {
		return "uninstall", nil
	} else if len(stateSignal.Agents.DatadogAgent.Version) > 0 {
		return "install", nil
	}
	return "", nil
}

// ValidateDocpInstalled return if state docp is installed
func (l *ManagerAdapter) ValidateDocpInstalled() (bool, error) {
	l.logger.Debug("validate docp installed", "trace", "docp-agent-os-instance.manager_adapter.ValidateDocpInstalled")
	content, err := l.GetStateReceived()
	if err != nil {
		return false, err
	}
	var agentState dto.StateCheckResponse
	if err := json.Unmarshal(content, &agentState); err != nil {
		return false, err
	}
	signal := agentState.Signal
	if len(signal.Agents.DocpAgent.Version) > 0 {
		status, err := l.osOperation.Status("agent")
		if err != nil {
			return false, err
		}
		if strings.ReplaceAll(status, "\"", "") == "active" {
			return true, nil
		}
	}
	return false, nil
}

// ValidateDocpNotInstalled return if state docp is not installed
func (l *ManagerAdapter) ValidateDocpNotInstalled() (bool, error) {
	l.logger.Debug("validate docp not installed", "trace", "docp-agent-os-instance.manager_adapter.ValidateDocpNotInstalled")
	content, err := l.GetStateReceived()
	if err != nil {
		return false, err
	}
	var agentState dto.StateCheckResponse
	if err := json.Unmarshal(content, &agentState); err != nil {
		return false, err
	}
	signal := agentState.Signal
	if len(signal.Agents.DocpAgent.Version) == 0 {
		status, err := l.osOperation.Status("agent")
		if err != nil {
			return false, err
		}
		if strings.ReplaceAll(status, "\"", "") == "active" {
			return true, nil
		}
	}
	return false, nil
}

// ValidateDatadogInstalled return if state datadog is installed
func (l *ManagerAdapter) ValidateDatadogInstalled() (bool, error) {
	l.logger.Debug("validate datadog installed", "trace", "docp-agent-os-instance.manager_adapter.ValidateDatadogInstalled")
	content, err := l.GetStateReceived()
	if err != nil {
		return false, err
	}
	var agentState dto.StateCheckResponse
	if err := json.Unmarshal(content, &agentState); err != nil {
		return false, err
	}
	signal := agentState.Signal
	if len(signal.Agents.DatadogAgent.Version) > 0 {
		status, err := l.osOperation.Status("datadog")
		if err != nil {
			return false, err
		}
		if strings.ReplaceAll(status, "\"", "") == "active" {
			return true, nil
		}
	}

	return false, nil
}

// ValidateDatadogNotInstalled return if state datadog is not installed
func (l *ManagerAdapter) ValidateDatadogNotInstalled() (bool, error) {
	l.logger.Debug("validate datadog not installed", "trace", "docp-agent-os-instance.manager_adapter.ValidateDatadogNotInstalled")
	content, err := l.GetStateReceived()
	if err != nil {
		return false, err
	}
	var agentState dto.StateCheckResponse
	if err := json.Unmarshal(content, &agentState); err != nil {
		return false, err
	}
	signal := agentState.Signal
	if len(signal.Agents.DatadogAgent.Version) == 0 {
		status, err := l.osOperation.Status("datadog")
		if err != nil {
			return false, err
		}
		if strings.ReplaceAll(status, "\"", "") == "active" {
			return true, nil
		}
	}
	return false, nil
}

// Validade execute validate states
func (l *ManagerAdapter) Validate() error {
	l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate")
	docpAgentIsOK := false
	datadogAgentIsOK := false

	workdir, err := utils.GetWorkDirPath()
	if err != nil {
		return err
	}

	receivedFilePath := filepath.Join(workdir, "state", "received")
	currentFilePath := filepath.Join(workdir, "state", "current")

	if err := l.fileSystem.VerifyFileExist(receivedFilePath); err != nil {
		return err
	}

	content, err := l.fileSystem.GetFileContent(receivedFilePath)
	if err != nil {
		return err
	}

	var agentState dto.StateCheckResponse
	if err := json.Unmarshal(content, &agentState); err != nil {
		return err
	}
	signal := agentState.Signal
	agent := signal.Agents.DocpAgent
	datadog := signal.Agents.DatadogAgent

	// validate docp agent
	l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate", "docpAgent", agent)
	if len(agent.Version) > 0 {
		status, err := l.osOperation.Status("agent")
		l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate", "docp agent status when enabled", status)
		if err != nil {
			return err
		}
		if strings.ReplaceAll(status, "\"", "") == "active" {
			docpAgentIsOK = true
			l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate", "docpAgentIsOK", docpAgentIsOK)
		}

	}

	if len(agent.Version) == 0 {
		status, err := l.osOperation.Status("agent")
		l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate", "docp agent status not when enabled", status, "docpAgentIsOk", docpAgentIsOK)
		if err != nil {
			return err
		}
		if strings.ReplaceAll(status, "\"", "") != "active" {
			docpAgentIsOK = true
		}

	}
	// validate datadog
	if len(datadog.Version) > 0 {
		status, err := l.osOperation.Status("datadog")
		if err != nil {
			return err
		}
		if strings.ReplaceAll(status, "\"", "") == "active" {
			datadogAgentIsOK = true
		}
	}
	if len(datadog.Version) == 0 {
		status, err := l.osOperation.Status("datadog")
		l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate", "datadog agent status", status)
		if err != nil {
			return err
		}
		if strings.ReplaceAll(status, "\"", "") != "active" {
			datadogAgentIsOK = true
		}
	}

	l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate", "docpAgentIsOk", docpAgentIsOK, "datadogAgentIsOk", datadogAgentIsOK)
	if docpAgentIsOK && datadogAgentIsOK {
		l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate", "status", "validation success")
		if err := l.fileSystem.WriteFileContent(currentFilePath, content); err != nil {
			return err
		}
		return nil
	}

	l.logger.Debug("validate", "trace", "docp-agent-os-instance.manager_adapter.Validate", "status", "validation failed")
	return nil
}

// CompareState execute compare states between received and current
func (l *ManagerAdapter) CompareState() (bool, error) {
	l.logger.Debug("compare state", "trace", "docp-agent-os-instance.manager_adapter.CompareState")
	receivedBytes, err := l.GetStateReceived()
	if err != nil {
		return false, err
	}
	currentBytes, err := l.GetStateCurrent()
	if err != nil {
		return false, err
	}
	md5ReceivedHash := utils.GenerateMd5Hash(receivedBytes)
	md5CurrentHash := utils.GenerateMd5Hash(currentBytes)
	l.logger.Debug("compare state", "trace", "docp-agent-os-instance.manager_adapter.CompareState", "md5 received hash", md5ReceivedHash, "md5 current hash", md5CurrentHash)
	if md5CurrentHash == md5ReceivedHash {
		return true, nil
	} else {
		return false, nil
	}
}

// IsAlreadyCreated return is already created host
func (l *ManagerAdapter) IsAlreadyCreated() (bool, error) {
	l.logger.Debug("is already created", "trace", "docp-agent-os-instance.manager_adapter.IsAlreadyCreated")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	content, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return false, err
	}
	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return false, err
	}
	if configAgent.AlreadyCreated {
		return true, nil
	}

	return false, nil
}

// SaveAlreadyTracer save already tracer on config file
func (l *ManagerAdapter) SaveAlreadyTracer(value bool) error {
	l.logger.Debug("save already tracer", "trace", "docp-agent-os-instance.manager_adapter.SaveAlreadyTracer")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return err
	}

	configAgent.AlreadyTracer = value

	newConfigAgentBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, newConfigAgentBytes); err != nil {
		return err
	}
	return nil
}

// GetAlreadyTracer return already tracer from config file
func (l *ManagerAdapter) GetAlreadyTracer() (bool, error) {
	l.logger.Debug("get already tracer", "trace", "docp-agent-os-instance.manager_adapter.GetAlreadyTracer")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return false, err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return false, err
	}

	return configAgent.AlreadyTracer, nil
}

// SaveInitialConfigFromRegister save initial config from register in file
func (l *ManagerAdapter) SaveInitialConfigFromRegister(data []byte) error {
	l.logger.Debug("save initial config from register", "trace", "docp-agent-os-instance.manager_adapter.SaveInitialConfigFromRegister", "data", string(data))
	var configAgent dto.ConfigAgent
	var registerDataResponseSuccess dto.AgentRegisterDataResponseSuccess
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	content, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return err
	}
	if err := l.ymlClient.Unmarshall(content, &configAgent); err != nil {
		return err
	}
	if err := l.unmarshaller(data, &registerDataResponseSuccess); err != nil {
		return err
	}
	token := registerDataResponseSuccess.AccessToken
	claims, err := utils.DecodeJwt(token)
	if err != nil {
		return err
	}
	configAgent.AccessToken = token
	configAgent.ComputeId = claims.ComputeId
	configAgent.DocpOrgId = claims.DocpOrgId
	configAgent.AlreadyCreated = true
	configAgent.AlreadyTracer = false

	newConfigAgentBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, newConfigAgentBytes); err != nil {
		return err
	}
	return nil
}

// ExistTracerLanguage verify if tracer language exist on host
func (l *ManagerAdapter) ExistTracerLanguage(language string) (bool, error) {
	l.logger.Debug("exist tracer language", "trace", "docp-agent-os-instance.manager_adapter.ExistTracerLanguage", "language", language)
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return false, err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return false, err
	}
	for _, lang := range configAgent.TracerLanguages {
		if lang == language {
			return true, nil
		}
	}

	return false, nil
}

// AddTracerLanguage append tracer language in slice the config file
func (l *ManagerAdapter) AddTracerLanguage(language string) error {
	l.logger.Debug("add tracer language", "trace", "docp-agent-os-instance.manager_adapter.AddTracerLanguage", "language", language)
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return err
	}
	slice := configAgent.TracerLanguages
	seen := make(map[string]struct{})

	for _, item := range slice {
		seen[item] = struct{}{}
	}

	if _, exists := seen[language]; !exists {
		slice = append(slice, language)
	}
	configAgent.TracerLanguages = slice

	bytYml, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, bytYml); err != nil {
		return err
	}

	return nil
}

// ClearTracerLanguage append tracer language in slice the config file
func (l *ManagerAdapter) ClearTracerLanguage() error {
	l.logger.Debug("clear tracer language", "trace", "docp-agent-os-instance.manager_adapter.ClearTracerLanguage")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return err
	}
	configAgent.TracerLanguages = []string{}

	bytYml, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, bytYml); err != nil {
		return err
	}

	return nil
}

// VerifyDatadogInstalled verify if datadog agent is installed
func (l *ManagerAdapter) VerifyDatadogInstalled() bool {
	pathDatadogBinaryAgent, err := utils.GetDatadogBinaryAgentPath()
	if err != nil {
		l.logger.Error("verify datadog installed", "error", err.Error())
		return false
	}
	info, err := os.Stat(pathDatadogBinaryAgent)
	if err != nil {
		l.logger.Error("verify datadog installed", "error", err.Error())
		return false
	}
	return info.IsDir()
}

// GetConfigAgent get config agent from file
func (l *ManagerAdapter) GetConfigAgent() (dto.ConfigAgent, error) {
	l.logger.Debug("get config agent", "trace", "docp-agent-os-instance.manager_adapter.GetConfigAgent")
	var configAgent dto.ConfigAgent
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")
	contentConfigAgent, err := l.fileSystem.GetFileContent(pathConfigFile)
	if err != nil {
		return dto.ConfigAgent{}, err
	}
	if err := l.ymlClient.Unmarshall(contentConfigAgent, &configAgent); err != nil {
		return dto.ConfigAgent{}, err
	}
	return configAgent, nil
}

// UpdateConfigAgent update config agent for file
func (l *ManagerAdapter) UpdateConfigAgent(configAgent dto.ConfigAgent) error {
	l.logger.Debug("update config agent", "trace", "docp-agent-os-instance.manager_adapter.UpdateConfigAgent")
	pathConfigFile := filepath.Join(l.agentWorkDir, "config.yml")

	configBytes, err := l.ymlClient.Marshall(&configAgent)
	if err != nil {
		return err
	}
	if err := l.fileSystem.WriteFileContent(pathConfigFile, configBytes); err != nil {
		return err
	}
	return nil
}

// GetStore return value from store
func (l *ManagerAdapter) GetStore(key string) any {
	return l.store.Get(key)
}

// SetStore execute set key and value to store
func (l *ManagerAdapter) SetStore(key string, val any) error {
	return l.store.Set(key, val)
}

// HandlerSCMManager execute handler for scm manager
func (l *ManagerAdapter) HandlerSCMManager() error {
	if err := l.osOperation.Execute(); err != nil {
		return err
	}
	return nil
}
