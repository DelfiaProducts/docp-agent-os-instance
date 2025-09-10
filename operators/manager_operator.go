package operators

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"

	adapters "github.com/DelfiaProducts/docp-agent-os-instance/libs/adapters"
	libdto "github.com/DelfiaProducts/docp-agent-os-instance/libs/dto"
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/pkg"
	services "github.com/DelfiaProducts/docp-agent-os-instance/libs/services"
	"github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
	libutils "github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

// ManagerOperator is struct for manager the operator
type ManagerOperator struct {
	filePath             string
	logger               libinterfaces.ILogger
	adapter              *adapters.ManagerAdapter
	vendorAdapter        *adapters.DatadogAdapter
	register             *services.AgentRegisterService
	stateCheck           *services.StateCheckService
	wg                   *sync.WaitGroup
	done                 chan struct{}
	chanErrors           chan dto.CommonChanErrors
	chanMetadata         chan []byte
	chanResultsApi       chan []byte
	chanDocpAgent        chan dto.StateAction
	chanDocpAgentDatadog chan dto.StateAction
	validateIsComplete   bool
	retryRegister        int
	maxRetry             int
	delay                time.Duration
}

// NewManagerOperator return instance of manager operator
func NewManagerOperator() *ManagerOperator {
	return &ManagerOperator{
		wg:                   &sync.WaitGroup{},
		done:                 make(chan struct{}),
		chanErrors:           make(chan dto.CommonChanErrors, 1),
		chanMetadata:         make(chan []byte, 1),
		chanResultsApi:       make(chan []byte, 1),
		chanDocpAgent:        make(chan dto.StateAction, 1),
		chanDocpAgentDatadog: make(chan dto.StateAction, 1),
		validateIsComplete:   false,
		retryRegister:        0,
		maxRetry:             10,
		delay:                time.Second * 1,
	}
}

// Setup configure operator
func (l *ManagerOperator) Setup() error {
	var logger libinterfaces.ILogger
	if runtime.GOOS == "windows" {
		workdir, err := libutils.GetWorkDirPath()
		if err != nil {
			return err
		}
		logPath := filepath.Join(workdir, "logs", "manager.log")
		loggerFile := libutils.NewDocpLoggerWindowsFileText(logPath)
		logger = loggerFile
	} else {
		logger = libutils.NewDocpLoggerJSON(os.Stdout)
	}
	l.logger = logger
	adapterManager := adapters.NewManagerAdapter(l.logger)
	if err := adapterManager.Prepare(); err != nil {
		return err
	}
	l.adapter = adapterManager
	vendorAdapter := adapters.NewDatadogAdapter(l.logger)
	if err := vendorAdapter.Setup(); err != nil {
		return err
	}
	l.vendorAdapter = vendorAdapter
	stateCheck := services.NewStateCheckService(l.logger)
	if err := stateCheck.Setup(); err != nil {
		return err
	}
	l.stateCheck = stateCheck
	serviceRegister := services.NewAgentRegisterService(l.logger)
	if err := serviceRegister.Setup(); err != nil {
		return err
	}
	l.register = serviceRegister

	filePath, err := libutils.GetConfigFilePath()
	if err != nil {
		return err
	}
	l.filePath = filePath
	return nil
}

// WaitGroupAdd add wait group
func (l *ManagerOperator) WaitGroupAdd(delta int) {
	l.wg.Add(delta)
}

// WaitGroupDone done wait group
func (l *ManagerOperator) WaitGroupDone() {
	l.wg.Done()
}

// WaitGroupDone execute wait
func (l *ManagerOperator) WaitGroupWait() {
	l.wg.Wait()
}

// GetState get state from state check
func (l *ManagerOperator) GetSignalFromStateCheck() error {
	l.logger.Debug("get signal from state check", "trace", "docp-agent-os-instance.manager_operator.GetSignalFromStateCheck")
	stateCheckBytes, statusCode, err := l.stateCheck.GetState()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "GetState", Priority: dto.ErrLevelMedium, Err: err}
	}
	l.logger.Debug("get state", "trace", "docp-agent-os-instance.manager_operator.GetState", "statusCode", statusCode)
	switch statusCode {
	case 200:
		l.validateIsComplete = false

		// validate if duplicated signal
		err := l.validateDuplicatedSignal(stateCheckBytes)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "GetSignalFromStateCheck", Priority: dto.ErrLevelMedium, Err: err}
			return err
		}
		// save signal on received state
		err = l.adapter.SaveState(stateCheckBytes)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "GetSignalFromStateCheck", Priority: dto.ErrLevelMedium, Err: err}
			return err
		}
	case 204:
		l.logger.Debug("get signal from state check", "trace", "docp-agent-os-instance.manager_operator.GetSignalFromStateCheck", "status", "not state present")
		return nil
	case 403:
		l.logger.Debug("get signal from state check", "trace", "docp-agent-os-instance.manager_operator.GetSignalFromStateCheck", "status", "not authorized")
		if err := l.executeAuthCall(); err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "GetSignalFromStateCheck", Priority: dto.ErrLevelMedium, Err: err}
			return err
		}
		return nil
	default:
		return nil
	}
	return nil
}

// UpdateAgent execute update the agent docp
func (l *ManagerOperator) UpdateAgent(version string) error {
	l.logger.Debug("update the agent docp", "trace", "docp-agent-os-instance.manager_operator.updateAgent")
	defer l.wg.Done()
	statusManager, err := l.adapter.Status("manager")
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
		return err
	}

	statusAgent, err := l.adapter.Status("agent")
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
		return err
	}

	if statusAgent != "active" {
		go l.installAgent()
		return nil
	}
	l.logger.Info("auto update agent version", "statusManager", statusManager, "statusAgent", statusAgent)
	if statusManager == "active" && statusAgent == "active" {
		agentVersion, err := l.adapter.GetAgentVersion()
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
			return err
		}
		l.logger.Info("auto update agent version", "version", version, "agentVersion", agentVersion)
		if version != agentVersion {

			transaction := libutils.NewTransactionStatus()
			ctx := context.WithValue(context.Background(), libdto.ContextTransactionStatus, transaction)

			go l.adapter.NotifyStatus("update_docp_received", pkg.TransactionEventOpen, "update docp received", ctx)
			time.Sleep(l.delay)

			go l.adapter.NotifyStatus("update_docp_initiate", pkg.TransactionEventUpdate, "update docp initialized", ctx)
			time.Sleep(l.delay)

			if err := l.adapter.UpdateAgent(version); err != nil {
				go l.adapter.NotifyStatus("update_docp_error", pkg.TransactionEventClose, "failed update agent version", ctx)
				l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
				return err
			}

			//save new version
			if err := l.adapter.SaveAgentVersion(version); err != nil {
				go l.adapter.NotifyStatus("update_docp_error", pkg.TransactionEventClose, "failed save agent version", ctx)
				l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
				return err
			}

			//save rollback version
			if err := l.adapter.SaveAgentRollbackVersion(agentVersion); err != nil {
				go l.adapter.NotifyStatus("update_docp_error", pkg.TransactionEventClose, "failed save agent version", ctx)
				l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
				return err
			}

			go l.adapter.NotifyStatus("update_docp_completed", pkg.TransactionEventClose, "update docp completed", ctx)
			return nil
		}
	}
	return nil
}

// GetActions get action for execute in manager
func (l *ManagerOperator) GetActions() ([]byte, error) {
	l.logger.Debug("get actions", "trace", "docp-agent-os-instance.manager_operator.GetActions")
	actions, err := l.adapter.GetState()
	l.logger.Debug("get actions", "trace", "docp-agent-os-instance.manager_operator.GetActions", "actions", string(actions))
	if err != nil {
		return nil, err
	}
	return actions, nil
}

// Stop execute stoping the ManagerOperator
func (l *ManagerOperator) Stop() {
	l.logger.Debug("stop", "trace", "docp-agent-os-instance.manager_operator.Stop")
	l.done <- struct{}{}
}

// AutoUpdateAgentVersion execute auto update agent version
func (l *ManagerOperator) AutoUpdateAgentVersion() error {
	received, err := l.adapter.GetStateReceived()
	if err != nil {
		return err
	}
	l.logger.Info("auto update agent version", "received", string(received))
	applyVersion, err := l.adapter.GetAgentVersionFromSignalBytes(received)
	if err != nil {
		return err
	}
	l.logger.Info("auto update agent version", "applyVersion", applyVersion)
	if applyVersion != "latest" {
		l.logger.Info("auto update agent version not latest", "applyVersion", applyVersion)
		return nil
	}
	//fetch agent versions
	agentVersions, err := l.adapter.FetchAgentVersions()
	if err != nil {
		return err
	}
	l.logger.Info("auto update agent version", "agentVersions", agentVersions)
	latestVersion := agentVersions.LatestVersion
	l.logger.Info("auto update agent version", "latestVersion", latestVersion)
	if len(latestVersion) > 0 {
		l.wg.Add(1)
		go l.UpdateAgent(latestVersion)
	}

	return nil
}

// updateAgentVersionDatadog execute update agent version datadog
func (l *ManagerOperator) UpdateAgentVersionDatadog(version string) error {
	l.logger.Debug("update agent version datadog", "trace", "docp-agent-os-instance.manager_operator.updateAgentVersionDatadog")
	defer l.wg.Done()

	executeRollback := false
	attempt := 0
	//get installed version datadog
	lastVersion, err := l.vendorAdapter.GetVersion()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "updateAgentVersionDatadog", Priority: dto.ErrLevelMedium, Err: err}
		return err
	}
	l.logger.Debug("installed version datadog", "lastVersion", lastVersion, "version", version)

	if lastVersion == version {
		l.logger.Debug("datadog agent version is already up to date")
		return nil
	}
	//prepare transaction
	transaction := utils.NewTransactionStatus()
	ctxTransaction := context.WithValue(context.Background(), dto.ContextTransactionStatus, transaction)

	go l.adapter.NotifyStatus("update_vendor_version_received", pkg.TransactionEventOpen, "update vendor version received", ctxTransaction)
	time.Sleep(l.delay)

	//send request for update version datadog
	_, err = l.adapter.DocpAgentApiUpdateVersionDatadog(version)
	if err != nil {
		go l.adapter.NotifyStatus("update_vendor_version_error", pkg.TransactionEventClose, "failed update vendor version", ctxTransaction)
		l.chanErrors <- dto.CommonChanErrors{From: "updateAgentVersionDatadog", Priority: dto.ErrLevelMedium, Err: err}
		return err
	}

	go l.adapter.NotifyStatus("update_vendor_version_processing", pkg.TransactionEventUpdate, "update vendor version processing", ctxTransaction)
	time.Sleep(l.delay)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

loopupdateversiondatadog:
	for {
		select {
		case <-ctx.Done():
			l.logger.Error("Context timeout or cancellation reached")
			l.chanErrors <- dto.CommonChanErrors{From: "updateAgentVersionDatadog", Priority: dto.ErrLevelMedium, Err: utils.ErrContextExpired()}
			go l.adapter.NotifyStatus("update_vendor_version_error", pkg.TransactionEventClose, "failed update vendor version", ctxTransaction)
			return utils.ErrContextExpired()

		case <-ticker.C:
			attempt++
			l.logger.Debug("checking datadog agent status", "attempt", attempt)
			datadogActive, err := l.adapter.Status("datadog")
			if err != nil {
				l.logger.Error("failed to get datadog agent service already installed", "error", err)
				l.chanErrors <- dto.CommonChanErrors{From: "updateAgentVersionDatadog", Priority: dto.ErrLevelMedium, Err: err}
				continue
			}

			if datadogActive == "active" {
				break loopupdateversiondatadog
			}

			if attempt >= l.maxRetry {
				l.logger.Error("max retry reached", "attempt", attempt)
				executeRollback = true
				break loopupdateversiondatadog
			}

		}
	}
	if executeRollback {
		l.logger.Debug("rollback update vendor version", "trace", "docp-agent-os-instance.manager_operator.UpdateAgentVersionDatadog")
		if err := l.vendorAdapter.RollbackVersion(lastVersion); err != nil {
			go l.adapter.NotifyStatus("update_vendor_version_rollback_error", pkg.TransactionEventClose, "failed update vendor version with rollback", ctxTransaction)
			l.chanErrors <- dto.CommonChanErrors{From: "updateAgentVersionDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return err
		}
		l.logger.Debug("rollback update vendor version", "lastVersion", lastVersion)
		go l.adapter.NotifyStatus("update_vendor_version_with_rollback_complete", pkg.TransactionEventClose, "update vendor version with rollback completed", ctxTransaction)
		return nil
	}
	l.logger.Debug("datadog agent service is already active")

	go l.adapter.NotifyStatus("update_vendor_version_complete", pkg.TransactionEventClose, "update vendor version completed", ctxTransaction)

	return nil
}

// AutoUpdateAgentDatadogVersion execute auto update agent datadog version
func (l *ManagerOperator) AutoUpdateAgentDatadogVersion() error {
	received, err := l.adapter.GetStateReceived()
	if err != nil {
		return err
	}
	l.logger.Info("auto update agent datadog version", "received", string(received))
	applyVersion, err := l.adapter.GetAgentVersionDatadogFromSignalBytes(received)
	if err != nil {
		return err
	}
	l.logger.Info("auto update agent datadog version", "applyVersion", applyVersion)
	if applyVersion != "latest" {
		l.logger.Info("auto update agent datadog version not latest", "applyVersion", applyVersion)
		return nil
	}
	//update repository
	if err := l.vendorAdapter.UpdateRepository(); err != nil {
		return err
	}

	//get latest and installed versions
	latestVersion, err := l.vendorAdapter.GetLatestVersion()
	if err != nil {
		return err
	}

	installedVersion, err := l.vendorAdapter.GetVersion()
	if err != nil {
		return err
	}

	if latestVersion == installedVersion {
		l.logger.Info("auto update agent datadog version already installed", "latestVersion", latestVersion, "installedVersion", installedVersion)
		return nil
	}

	if err := l.UpdateAgentVersionDatadog(latestVersion); err != nil {
		return err
	}

	return nil
}

// Run execute loop the manager
func (l *ManagerOperator) Run() error {
	if err := l.Setup(); err != nil {
		return err
	}
	l.logger.Info("execute manager")
	l.logger.Debug("execute running", "trace", "docp-agent-os-instance.manager_operator.Run")
	defer l.logger.Close()
	l.wg.Add(14)
	go l.persistLastSignalHashToStore()
	go l.comunicateSCM()
	go l.Profiling()
	go l.Start()
	go l.consumerErrors()
	go l.consumerResultsFromApiDocpAgent()
	go l.collectGetState()
	go l.collectGetActions()
	go l.consumeAllActions()
	go l.handleMetadata()
	go l.periodicHandlerMetadata()
	go l.periodicAutoUpdate()
	go l.periodicTasks()
	go l.getMetadata()
	l.wg.Wait()
	return nil
}

// Start execute mathod for running in manager operator
func (l *ManagerOperator) Start() {
	l.logger.Debug("start tasks", "trace", "docp-agent-os-instance.manager_operator.Start")
	defer l.wg.Done()

	l.wg.Add(1)
	go l.installAgent()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.wg.Add(4)
			go l.collectGetState()
			go l.validateDocpAgentInstalled()
			go l.validateDatadogAgentInstalled()
			go l.validateDatadogAgentUpdateConfigs()

		case <-l.done:
			close(l.done)
			return
		}
	}
}

func (l *ManagerOperator) CheckHealth() {
	l.logger.Debug("in CheckHealth", "trace", "docp-agent-os-instance.manager_operator.CheckHealth")
	resp, err := http.Get("http://localhost:3000/health")
	if err != nil {
		l.logger.Error("in CheckHealth", "trace", "docp-agent-os-instance.manager_operator.CheckHealth", "error", err.Error())
		return
	}
	defer resp.Body.Close()

	var healthResp dto.HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	if err != nil {
		l.logger.Error("in CheckHealth", "trace", "docp-agent-os-instance.manager_operator.CheckHealth", "error", err.Error())
		return
	}

	if healthResp.Status == "success" && healthResp.Code == "HEALTH_OK" {
		l.logger.Debug("check health success", "trace", "docp-agent-os-instance.manager_operator.CheckHealth", "healthResp", healthResp)
	} else {
		l.logger.Warn("check health failed", "trace", "docp-agent-os-instance.manager_operator.CheckHealth", "healthResp", healthResp)
	}
}

func (l *ManagerOperator) Profiling() {
	l.logger.Debug("profiling task", "trace", "docp-agent-os-instance.manager_operator.Profiling")
	defer l.wg.Done()
	if err := http.ListenAndServe(":4040", nil); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "Profiling", Priority: dto.ErrLevelLow, Err: err}
	}
}

// Close execute close the operator loop
func (l *ManagerOperator) Close() error {
	l.logger.Debug("execute close", "trace", "docp-agent-os-instance.linux_manager_operator.Close")
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		l.adapter.Close()
	}()
	return nil
}

// marshaller execute marshal the inner
func (l *ManagerOperator) marshaller(inner any) ([]byte, error) {
	l.logger.Debug("marshaller", "trace", "docp-agent-os-instance.manager_operator.marshaller", "inner", inner)
	slcBytes, err := json.Marshal(inner)
	if err != nil {
		return nil, err
	}
	return slcBytes, nil
}

// unmarshaller execute unmarshal the inner
func (l *ManagerOperator) unmarshaller(content []byte, inner any) error {
	l.logger.Debug("unmarshaller", "trace", "docp-agent-os-instance.manager_operator.unmarshaller", "content", string(content), "inner", inner)
	if err := json.Unmarshal(content, inner); err != nil {
		return err
	}
	return nil
}

func (l *ManagerOperator) getLevelError() int {
	l.logger.Debug("get level error", "trace", "docp-agent-os-instance.manager_operator.getLevelError")
	envLevelError := os.Getenv("ERROR_LEVEL")
	switch envLevelError {
	case "high":
		return dto.ErrLevelHigh
	case "medium":
		return dto.ErrLevelMedium
	case "low":
		return dto.ErrLevelLow
	default:
		return dto.ErrLevelHigh
	}
}

// getVendorLatestVersion retrieves the latest version of the vendor
func (l *ManagerOperator) getVendorLatestVersion() (string, error) {
	l.logger.Debug("get vendor latest version", "trace", "docp-agent-os-instance.manager_operator.getVendorLatestVersion")
	latestVersion, err := l.vendorAdapter.GetLatestVersion()
	if err != nil {
		return "", err
	}
	return latestVersion, nil
}

// consumerErrors execute consumer for errors
func (l *ManagerOperator) consumerErrors() {
	l.logger.Debug("execute consumer errors", "trace", "docp-agent-os-instance.manager_operator.consumerErrors")
	defer l.wg.Done()
	for {
		select {
		case managerErr, ok := <-l.chanErrors:
			if !ok {
				return
			}
			if managerErr.Err != nil && managerErr.Priority <= l.getLevelError() {
				l.logger.Error("error received in consumer errors", "from", managerErr.From, "error", managerErr.Err.Error())
			}
		}
	}
}

// resolveAgentVersion resolve the agent version
func (l *ManagerOperator) resolveAgentVersion() (string, error) {
	l.logger.Debug("resolve agent version", "trace", "docp-agent-os-instance.manager_operator.resolveAgentVersion")
	var version string
	versionInstalled, err := l.adapter.GetAgentVersion()
	if err != nil {
		return version, err
	}
	version = versionInstalled
	if version == "latest" {
		agentVersions, err := l.adapter.FetchAgentVersions()
		if err != nil {
			return version, err
		}
		version = agentVersions.LatestVersion
	}
	return version, nil
}

// executeAuthCall execute call to auth and save access token received
func (l *ManagerOperator) executeAuthCall() error {
	if err := l.adapter.ExecuteAuthCall(); err != nil {
		return err
	}

	return nil
}

// retryHandlerMetadata execute retry the create initial data
func (l *ManagerOperator) retryHandlerMetadata() error {
	l.logger.Debug("retry handler register", "timestamp", time.Now())
	l.retryRegister += 1
	time.Sleep(time.Minute * time.Duration(l.retryRegister))
	l.handleMetadata()
	return nil
}

// handleMetadata execute send metadata to register
func (l *ManagerOperator) handleMetadata() {
	l.logger.Debug("execute handle metadata", "trace", "docp-agent-os-instance.linux_manager_operator.handleMetadata")
	defer l.wg.Done()
	isAlreadyCreated, err := l.adapter.IsAlreadyCreated()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "handleMetadata", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	if isAlreadyCreated {
		l.wg.Add(1)
		go l.sendMetadataUpdate()
		return
	} else {
		l.wg.Add(1)
		go l.sendMetadataCreate()
		return
	}
}

// sendMetadataCreate execute send initial metadata to register
func (l *ManagerOperator) sendMetadataCreate() {
	l.logger.Debug("execute send metadata create", "trace", "docp-agent-os-instance.linux_manager_operator.sendMetadataCreate")
	defer l.wg.Done()
	for {
		select {
		case metadata, ok := <-l.chanMetadata:
			if !ok {
				return
			}
			result, statusCode, err := l.register.SendMetadataCreate(metadata)
			if err != nil {
				l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataCreate", Priority: dto.ErrLevelMedium, Err: err}
			}

			l.logger.Debug("execute send metadata create", "trace", "docp-agent-os-instance.linux_manager_operator.sendMetadataCreate", "statusCode", statusCode)
			switch statusCode {
			case 202:
				if err := l.adapter.SaveInitialConfigFromRegister(result); err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataCreate", Priority: dto.ErrLevelMedium, Err: err}
					return
				}
			default:
				l.logger.Info("result from register service", "trace", "docp-agent-os-instance.linux_manager_operator.sendMetadata", "result", string(result))
				if l.retryRegister <= l.maxRetry {
					go l.retryHandlerMetadata()
				}
			}
		}
	}
}

// sendMetadataUpdate execute send update metadata to register
func (l *ManagerOperator) sendMetadataUpdate() {
	l.logger.Debug("execute send metadata update", "trace", "docp-agent-os-instance.linux_manager_operator.sendMetadataUpdate")
	defer l.wg.Done()
	for {
		select {
		case metadata, ok := <-l.chanMetadata:
			if !ok {
				return
			}

			transaction := libutils.NewTransactionStatus()
			ctx := context.WithValue(context.Background(), libdto.ContextTransactionStatus, transaction)

			go l.adapter.NotifyStatus("update_metadata", pkg.TransactionEventOpen, "update metadata", ctx)

			result, statusCode, err := l.register.SendMetadataUpdate(metadata)
			if err != nil {
				go l.adapter.NotifyStatus("update_metadata_error", pkg.TransactionEventClose, "error on update metadata", ctx)
				l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataUpdate", Priority: dto.ErrLevelMedium, Err: err}
			}

			go l.adapter.NotifyStatus("update_metadata_completed", pkg.TransactionEventClose, "update metadata completed", ctx)

			l.logger.Debug("execute send metadata update", "trace", "docp-agent-os-instance.linux_manager_operator.sendMetadataUpdate", "statusCode", statusCode)
			switch statusCode {
			case 202:
				configAgent, err := l.adapter.GetConfigAgent()
				if err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataUpdate", Priority: dto.ErrLevelMedium, Err: err}
				}

				var agentRegisterResponse libdto.AgentRegisterDataResponseSuccess
				if err := l.unmarshaller(result, &agentRegisterResponse); err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataUpdate", Priority: dto.ErrLevelMedium, Err: err}
				}

				if len(agentRegisterResponse.AccessToken) > 0 {
					configAgent.AccessToken = agentRegisterResponse.AccessToken
				}

				if err := l.adapter.UpdateConfigAgent(configAgent); err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataUpdate", Priority: dto.ErrLevelMedium, Err: err}
				}

			case 401:
				if err := l.executeAuthCall(); err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataUpdate", Priority: dto.ErrLevelMedium, Err: err}
				}

				if l.retryRegister <= l.maxRetry {
					go l.retryHandlerMetadata()
				}
			default:
				l.logger.Info("result from register service", "trace", "docp-agent-os-instance.linux_manager_operator.sendMetadata", "result", string(result))
			}
		}
	}
}

// validateDuplicatedSignal execute validate signal duplicated
// and notify transaction error with reason
func (l *ManagerOperator) validateDuplicatedSignal(signalBytes []byte) error {
	l.logger.Debug("validate duplicated signal", "trace", "docp-agent-os-instance.manager_operator.validateDuplicatedSignal")
	receivedBytes, err := l.adapter.GetStateReceived()
	if err != nil {
		return err
	}
	signalHash := utils.GenerateMd5Hash(signalBytes)
	receivedHash := utils.GenerateMd5Hash(receivedBytes)

	// validate if states received e signal received is equals
	// send notify if equals
	if signalHash == receivedHash {
		l.logger.Debug("validate duplicated signal", "equals", true)
		transaction := utils.NewTransactionStatus()
		ctxTransaction := context.WithValue(context.Background(), libdto.ContextTransactionStatus, transaction)
		go l.adapter.NotifyStatus("state_received", pkg.TransactionEventOpen, "verify state already exists", ctxTransaction)
		time.Sleep(l.delay)
		go l.adapter.NotifyStatus("state_completed", pkg.TransactionEventClose, "state already exists", ctxTransaction)
		return utils.ErrSignalAlreadyExists()
	}
	return nil
}

// getMetadata return metadata from host
func (l *ManagerOperator) getMetadata() {
	l.logger.Debug("execute get metadata", "trace", "docp-agent-os-instance.linux_manager_operator.getMetadata")
	for metadata := range l.adapter.Collect() {
		isChangedMetadata := l.verifyChangeMetadata(metadata)
		l.logger.Debug("execute get metadata", "trace", "docp-agent-os-instance.linux_manager_operator.getMetadata", "isChangedMetadata", isChangedMetadata)
		if isChangedMetadata {
			l.chanMetadata <- metadata
		}
	}
	close(l.chanMetadata)
	l.wg.Done()
}

// extractDDApiKeyAndDDSiteFromEnvs return envs for install datadog agent
func (l *ManagerOperator) extractDDApiKeyAndDDSiteFromEnvs(envs []dto.StateActionEnvs) (string, string, error) {
	l.logger.Debug("extract envs the datadog", "trace", "docp-agent-os-instance.manager_operator.extractDDApiKeyAndDDSiteFromEnvs", "envs", envs)
	var ddApiKey string
	var ddSite string
	for _, env := range envs {
		if env.Name == "DD_API_KEY" {
			ddApiKey = env.Value
		}
		if env.Name == "DD_SITE" {
			ddSite = env.Value
		}
	}
	return ddApiKey, ddSite, nil
}

// extractDDApiKeyAndDDSiteFromEnvs return envs for install datadog agent
func (l *ManagerOperator) extractApmSingleStepEnvs(envs []dto.StateActionEnvs) (string, string, string, error) {
	l.logger.Debug("extract envs the datadog", "trace", "docp-agent-os-instance.manager_operator.extractDDApiKeyAndDDSiteFromEnvs", "envs", envs)
	var ddApmInstrumentationEnabled string
	var ddEnv string
	var ddApmInstrumentationLibraries string
	for _, env := range envs {
		if env.Name == "DD_APM_INSTRUMENTATION_ENABLED" {
			ddApmInstrumentationEnabled = env.Value
		}
		if env.Name == "DD_ENV" {
			ddEnv = env.Value
		}
		if env.Name == "DD_APM_INSTRUMENTATION_LIBRARIES" {
			ddApmInstrumentationLibraries = env.Value
		}
	}
	if len(ddApmInstrumentationEnabled) == 0 {
		return "", "", "", errors.New("invalid dd apm instrumentation enabled")
	}
	if len(ddApmInstrumentationLibraries) == 0 {
		return "", "", "", errors.New("invalid dd apm instrumentation libraries")
	}
	return ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries, nil
}

// extractApmTracingLibrayEnvs return envs for install datadog agent
func (l *ManagerOperator) extractApmTracingLibrayEnvs(envs []dto.StateActionEnvs) (string, string, string, error) {
	l.logger.Debug("extract envs the datadog", "trace", "docp-agent-os-instance.manager_operator.extractApmTracingLibrayEnvs", "envs", envs)
	var language, pathTracer, version string
	for _, env := range envs {

		if env.Name == "language" {
			language = env.Value
		}
		if env.Name == "path_tracer" {
			pathTracer = env.Value
		}
		if env.Name == "version" {
			version = env.Value
		}
	}
	if len(language) == 0 {
		return "", "", "", errors.New("invalid dd apm language tracing library")
	}
	return language, pathTracer, version, nil
}

// installAgent execute install the agent docp
func (l *ManagerOperator) installAgent() {
	l.logger.Debug("install the agent docp", "trace", "docp-agent-os-instance.manager_operator.installAgent")
	defer l.wg.Done()
	status, err := l.adapter.Status("agent")
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "installAgent", Priority: dto.ErrLevelHigh, Err: err}
		return
	}
	if status != "active" {
		//get version
		version, err := l.adapter.GetAgentVersion()
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "installAgent", Priority: dto.ErrLevelHigh, Err: err}
			return
		}
		if err := l.adapter.InstallAgent(version); err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "installAgent", Priority: dto.ErrLevelHigh, Err: err}
			return
		}
		return
	}
	return
}

// installAgentDatadog execute call to api docp agent
// to install datadog agent
func (l *ManagerOperator) installAgentDatadog(ddApiKey, ddSite string) {
	l.logger.Debug("install agent datadog", "trace", "docp-agent-os-instance.manager_operator.installAgentDatadog", "ddApiKey", ddApiKey, "ddSite", ddSite)
	defer l.wg.Done()
	status, err := l.adapter.Status("datadog")
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "installAgentDatadog", Priority: dto.ErrLevelHigh, Err: err}
		return
	}
	if status != "active" {
		result, err := l.adapter.DocpAgentApiInstallDatadog(ddApiKey, ddSite)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "installAgentDatadog", Priority: dto.ErrLevelHigh, Err: err}
			return
		}
		l.chanResultsApi <- result
		return
	}
}

// handlerUpdateAgentDatadogAfterInstall execute install agent datadog and update configurations after agent active
func (l *ManagerOperator) handlerUpdateAgentDatadogAfterInstall(files []dto.StateActionFiles) {
	l.logger.Debug("install and update agent datadog", "trace", "docp-agent-os-instance.manager_operator.installAndUpdateAgentDatadog")
	defer l.wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

loopinstalldatadog:
	for {
		select {
		case <-ctx.Done():
			l.logger.Error("Context timeout or cancellation reached")
			l.chanErrors <- dto.CommonChanErrors{From: "installAndUpdateAgentDatadog", Priority: dto.ErrLevelMedium, Err: fmt.Errorf("installation process timed out")}
			return

		case <-ticker.C:
			datadogActive, err := l.adapter.Status("datadog")
			if err != nil {
				l.logger.Error("failed to get datadog agent service already installed", "error", err)
				l.chanErrors <- dto.CommonChanErrors{From: "installAndUpdateAgentDatadog", Priority: dto.ErrLevelMedium, Err: err}
				continue
			}

			if datadogActive == "active" {
				break loopinstalldatadog
			}

			l.logger.Debug("datadog agent service is already active")
		}
	}

	// delay for datadog agent configure all files terminated
	time.Sleep(time.Minute * 1)

	// update configurations datadog
	for _, fls := range files {
		flsBytes, err := l.marshaller(&fls)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return
		}

		l.wg.Add(1)
		go l.updateAgentDatadog(flsBytes)
	}
}

func (l *ManagerOperator) handlerInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries string) {
	l.logger.Debug("handle install datadog agent with APM single step", "trace", "docp-agent-os-instance.manager_operator.handlerInstallDatadogWithApmSingleStep", "ddApiKey", ddApiKey, "ddSite", ddSite, "ddApmInstrumentationEnabled", ddApmInstrumentationEnabled, "ddEnv", ddEnv, "ddApmInstrumentationLibraries", ddApmInstrumentationLibraries)
	defer l.wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	l.wg.Add(1)
	go l.uninstallAgentDatadog()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.logger.Error("Context timeout or cancellation reached")
			l.chanErrors <- dto.CommonChanErrors{From: "handlerInstallDatadogWithApmSingleStep", Priority: dto.ErrLevelMedium, Err: fmt.Errorf("installation process timed out")}
			return

		case <-ticker.C:
			alreadyInstalled, err := l.adapter.AlreadyInstalled("datadog")
			if err != nil {
				l.logger.Error("Failed to get Datadog agent service already installed", "error", err)
				l.chanErrors <- dto.CommonChanErrors{From: "handlerInstallDatadogWithApmSingleStep", Priority: dto.ErrLevelMedium, Err: err}
				continue
			}

			l.logger.Info("Datadog agent service alreadyInstalled", "alreadyInstalled", alreadyInstalled)

			if !alreadyInstalled {
				result, err := l.adapter.DocpAgentApiInstallDatadogWithApmSingleStep(
					ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries,
				)
				if err != nil {
					l.logger.Error("Failed to install Datadog agent with APM single step", "error", err)
					l.chanErrors <- dto.CommonChanErrors{From: "handlerInstallDatadogWithApmSingleStep", Priority: dto.ErrLevelMedium, Err: err}
					return
				}

				l.logger.Debug("Datadog agent installed successfully")
				l.chanResultsApi <- result
				return
			}

			l.logger.Debug("Datadog agent service is already active")
		}
	}
}

// installAgentDatadog execute call to api docp agent
// to install datadog agent
func (l *ManagerOperator) installAgentDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries string) {
	l.logger.Debug("install agent datadog with apm single step", "trace", "docp-agent-os-instance.manager_operator.installAgentDatadogWithApmSingleStep", "ddApiKey", ddApiKey, "ddSite", ddSite, "ddApmInstrumentationEnabled", ddApmInstrumentationEnabled, "ddEnv", ddEnv, "ddApmInstrumentationLibraries", ddApmInstrumentationLibraries)
	defer l.wg.Done()
	status, err := l.adapter.Status("datadog")
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "installAgentDatadogWithApmSingleStep", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	alreadyTracer, err := l.adapter.GetAlreadyTracer()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "installAgentDatadogWithApmSingleStep", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	l.logger.Debug("install agent datadog with apm single step", "trace", "docp-agent-os-instance.manager_operator.installAgentDatadogWithApmSingleStep", "docpStatus", status, "alreadyTracer", alreadyTracer)
	if status != "active" {
		result, err := l.adapter.DocpAgentApiInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "installAgentDatadogWithApmSingleStep", Priority: dto.ErrLevelMedium, Err: err}
			return
		}
		l.chanResultsApi <- result
		return
	} else if status == "active" && !alreadyTracer {
		l.wg.Add(1)
		go l.handlerInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries)
		return
	}
}

// installDatadogTracerWithTracingLibrary execute call to api docp agent
// to install datadog tracer with tracing library
func (l *ManagerOperator) installDatadogTracerWithTracingLibrary(ddApiKey, ddSite, language, pathTracer, version string) {
	l.logger.Debug("install datadog tracer", "trace", "docp-agent-os-instance.manager_operator.installDatadogTracerWithTracingLibrary", "ddApiKey", ddApiKey, "ddSite", ddSite, "language", language, "pathTracer", pathTracer, "version", version)
	defer l.wg.Done()
	status, err := l.adapter.Status("datadog")
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "installDatadogTracerWithTracingLibrary", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	alreadyTracer, err := l.adapter.GetAlreadyTracer()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "installDatadogTracerWithTracingLibrary", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	l.logger.Debug("install datadog tracer", "trace", "docp-agent-os-instance.manager_operator.installDatadogTracerWithTracingLibrary", "docpStatus", status, "alreadyTracer", alreadyTracer)
	existLanguage, err := l.adapter.ExistTracerLanguage(language)
	if err != nil {
		return
	}
	l.logger.Debug("install datadog tracer", "trace", "docp-agent-os-instance.manager_operator.installDatadogTracerWithTracingLibrary", "language", language, "existLanguage", existLanguage)
	if status == "active" && !existLanguage {
		if err := l.adapter.AddTracerLanguage(language); err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "installDatadogTracerWithTracingLibrary", Priority: dto.ErrLevelMedium, Err: err}
			return
		}
		l.wg.Add(1)
		go l.adapter.DocpAgentApiInstallDatadogWithApmTracingLibrary(ddApiKey, ddSite, language, pathTracer, version)
		return
	}
}

// uninstallAgentDatadog exeuct call to api docp agent
// to uninstall datadog agent
func (l *ManagerOperator) uninstallAgentDatadog() {
	l.logger.Debug("uninstall agent datadog", "trace", "docp-agent-os-instance.manager_operator.uninstallAgentDatadog")
	defer l.wg.Done()

	result, err := l.adapter.DocpAgentApiUninstallDatadog()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "uninstallAgentDatadog", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	l.chanResultsApi <- result
	if err := l.adapter.ClearTracerLanguage(); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "uninstallAgentDatadog", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	return
}

// updateAgentDatadog execute call to api docp agent
// to update datadog agent
func (l *ManagerOperator) updateAgentDatadog(content []byte) {
	l.logger.Debug("update agent datadog", "trace", "docp-agent-os-instance.manager_operator.updateAgentDatadog", "content", string(content))
	defer l.wg.Done()

	newHash := utils.GenerateMd5Hash(content)
	existsDatadogHash := l.adapter.GetStore("update.datadog.hash")

	if existsDatadogHash == nil {
		result, err := l.adapter.DocpAgentApiUpdateConfigurationsDatadog(content)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "updateAgentDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return
		}

		l.chanResultsApi <- result
		if err := l.adapter.SetStore("update.datadog.hash", newHash); err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "updateAgentDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return
		}

		if err := l.adapter.DaemonReload(); err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "updateAgentDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return
		}
		if err := l.adapter.RestartService("datadog"); err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "updateAgentDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return
		}
		return
	}

	return
}

// autoUninstallWithOtherVendors execute auto uninstall with vendors
func (l *ManagerOperator) autoUninstallWithOtherVendors() {
	l.logger.Debug("auto uninstall with other vendors the manager", "trace", "docp-agent-os-instance.manager_operator.autoUninstallWithOtherVendors")
	defer l.wg.Done()

	removeDatadog := false
	vendorsRemoved := make(map[string]bool)

	allVendors, err := l.adapter.GetRemoveOtherVendors()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "autoUninstallWithOtherVendors", Priority: dto.ErrLevelMedium, Err: err}
		return
	}

	removeOtherVendors := utils.RemoveItemFromSlice(allVendors, "all")

	for _, vendor := range removeOtherVendors {
		if vendor == "datadog" {
			removeDatadog = true
		}
	}

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
loopuninstall:
	for {
		select {
		case <-ticker.C:
			l.logger.Debug("auto uninstall with other vendors the manager", "vendors", removeOtherVendors)
			// validate and remove agent datadog
			if removeDatadog {

				statusDatadog, err := l.adapter.Status("datadog")
				if err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "autoUninstallWithOtherVendors", Priority: dto.ErrLevelMedium, Err: err}
					return
				}

				l.logger.Debug("auto uninstall with other vendors the manager", "statusDatadog", statusDatadog)
				if statusDatadog != "active" {
					ok := vendorsRemoved["datadog"]
					if !ok {
						vendorsRemoved["datadog"] = false
					}
					vendorsRemoved["datadog"] = true
				}
			}

			l.logger.Debug("auto uninstall with other vendors the manager", "lenght removeOtherVendors", len(removeOtherVendors), "lenght vendorsRemoved", len(vendorsRemoved))
			// validate all vendors uninstalled
			if len(removeOtherVendors) == len(vendorsRemoved) {
				break loopuninstall
			}
		case <-l.done:
			close(l.done)
			return
		}
	}

	// execute autouninstall for docp agent
	transaction := libutils.NewTransactionStatus()
	ctx := context.WithValue(context.Background(), libdto.ContextTransactionStatus, transaction)

	go l.adapter.NotifyStatus("uninstall_docp_received", pkg.TransactionEventOpen, "uninstall docp received", ctx)
	time.Sleep(l.delay)

	go l.adapter.NotifyStatus("uninstall_docp_processing", pkg.TransactionEventUpdate, "uninstall docp processing", ctx)
	time.Sleep(l.delay)

	//get version installed
	version, err := l.resolveAgentVersion()
	if err != nil {
		go l.adapter.NotifyStatus("uninstall_docp_error", pkg.TransactionEventClose, "failed uninstall docp", ctx)
		l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
		return
	}

	if err := l.adapter.AutoUninstall(version); err != nil {
		go l.adapter.NotifyStatus("uninstall_docp_error", pkg.TransactionEventClose, "failed uninstall docp", ctx)
		l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
		return
	}
	go l.adapter.NotifyStatus("uninstall_docp_completed", pkg.TransactionEventClose, "uninstall docp completed", ctx)

	return
}

// autoUninstallAgent execute auto uninstall the manager
func (l *ManagerOperator) autoUninstall() {
	l.logger.Debug("auto uninstall the manager", "trace", "docp-agent-os-instance.manager_operator.autoUninstall")

	// validate if exists other vendors and execute autoUninstallWithOtherVendors
	existsOtherVendor, err := l.adapter.ExisteOtherVendors()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
	}
	if existsOtherVendor {
		l.wg.Add(1)
		go l.autoUninstallWithOtherVendors()
	} else {
		transaction := libutils.NewTransactionStatus()
		ctx := context.WithValue(context.Background(), libdto.ContextTransactionStatus, transaction)

		go l.adapter.NotifyStatus("uninstall_docp_received", pkg.TransactionEventOpen, "uninstall docp received", ctx)
		time.Sleep(l.delay)

		defer l.wg.Done()

		go l.adapter.NotifyStatus("uninstall_docp_processing", pkg.TransactionEventUpdate, "uninstall docp processing", ctx)
		time.Sleep(l.delay)

		//get version
		version, err := l.resolveAgentVersion()
		if err != nil {
			go l.adapter.NotifyStatus("uninstall_docp_error", pkg.TransactionEventClose, "failed uninstall docp", ctx)
			l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
			return
		}

		if err := l.adapter.AutoUninstall(version); err != nil {
			go l.adapter.NotifyStatus("uninstall_docp_error", pkg.TransactionEventClose, "failed uninstall docp", ctx)
			l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
			return
		}

		go l.adapter.NotifyStatus("uninstall_docp_completed", pkg.TransactionEventClose, "uninstall docp completed", ctx)
	}
	return
}

// consumerResultsFromApiDocpAgent execute consume the result
// from api docp agent
func (l *ManagerOperator) consumerResultsFromApiDocpAgent() {
	l.logger.Debug("consumer results from api docp agent", "trace", "docp-agent-os-instance.manager_operator.consumerResultsFromApiDocpAgent")
	defer l.wg.Done()
	for res := range l.chanResultsApi {
		l.logger.Debug("consumer results from api docp agent", "trace", "docp-agent-os-instance.manager_operator.consumerResultsFromApiDocpAgent", "result", string(res))
	}
}

// consumeActionsDocpAgent execute consume the actions the agent
func (l *ManagerOperator) consumeActionsDocpAgent() {
	l.logger.Debug("consume actions docp agent", "trace", "docp-agent-os-instance.manager_operator.consumeActionsDocpAgent")
	defer l.wg.Done()
	for act := range l.chanDocpAgent {
		if act.Action == "update" {
			l.wg.Add(1)
			go l.UpdateAgent(act.Version)
		} else if act.Action == "uninstall" {
			l.wg.Add(1)
			go l.autoUninstall()
		} else {
			continue
		}
	}
}

// consumerActionsDatadog execute consume the actions
// for datadog agent
func (l *ManagerOperator) consumerActionsDatadog() {
	l.logger.Debug("consumer actions datadog", "trace", "docp-agent-os-instance.manager_operator.consumerActionsDatadog")
	defer l.wg.Done()

	// get actions for datadog
	for act := range l.chanDocpAgentDatadog {
		// verify if datadog already installed
		datadogAlreadyInstalled, err := l.adapter.AlreadyInstalled("datadog")
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return
		}
		l.logger.Debug("consumer actions datadog", "trace", "docp-agent-os-instance.manager_operator.consumerActionsDatadog", "action", act)
		// action update configurations datadog
		if act.Action == "update" {
			for _, fls := range act.Files {
				flsBytes, err := l.marshaller(&fls)
				if err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
					return
				}

				// if agent already installed execute update configurations
				if datadogAlreadyInstalled {
					l.wg.Add(2)
					go l.updateAgentDatadog(flsBytes)
					//validate if version is latest
					if act.Version == "latest" {
						latestVersion, err := l.getVendorLatestVersion()
						if err != nil {
							l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
							return
						}
						act.Version = latestVersion
					}
					//dispatch update version
					go l.UpdateAgentVersionDatadog(act.Version)
				}
			}
		}

		l.logger.Debug("consumer actions datadog", "trace", "docp-agent-os-instance.manager_operator.consumerActionsDatadog", "datadogAlreadyInstalled", datadogAlreadyInstalled)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
		}
		if act.Action == "install" {
			ddApiKey, ddSite, err := l.extractDDApiKeyAndDDSiteFromEnvs(act.Envs)
			if err != nil {
				l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
				return
			}
			if act.Component == "tracer" {
				if act.Mode == "single_step" {
					if datadogAlreadyInstalled {
						ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries, err := l.extractApmSingleStepEnvs(act.ComponentEnvs)
						if err != nil {
							l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
							return
						}
						l.wg.Add(1)
						go l.handlerInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries)
					} else {
						l.wg.Add(1)
						ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries, err := l.extractApmSingleStepEnvs(act.ComponentEnvs)
						if err != nil {
							l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
							return
						}
						go l.installAgentDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationEnabled, ddEnv, ddApmInstrumentationLibraries)
					}
				} else if act.Mode == "tracing_library" {
					if datadogAlreadyInstalled {
						language, pathTracer, version, err := l.extractApmTracingLibrayEnvs(act.ComponentEnvs)
						if err != nil {
							l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
							return
						}
						l.wg.Add(1)
						go l.installDatadogTracerWithTracingLibrary(ddApiKey, ddSite, language, pathTracer, version)
					}
				}
			} else if act.Component == "agent" {
				if !datadogAlreadyInstalled {
					if len(act.Files) > 0 {
						l.wg.Add(2)
						go l.installAgentDatadog(ddApiKey, ddSite)
						go l.handlerUpdateAgentDatadogAfterInstall(act.Files)
					} else {
						l.wg.Add(1)
						go l.installAgentDatadog(ddApiKey, ddSite)
					}
				}
			}

		} else if act.Action == "uninstall" {
			if datadogAlreadyInstalled {
				l.wg.Add(1)
				go l.uninstallAgentDatadog()
			}
		} else {
			continue
		}
	}
}

func (l *ManagerOperator) consumeAllActions() {
	l.logger.Debug("consume all actions", "trace", "docp-agent-os-instance.manager_operator.consumeAllActions")
	defer l.wg.Done()
	l.wg.Add(2)
	go l.consumeActionsDocpAgent()
	go l.consumerActionsDatadog()
}

// managerActions execut segment actions by type
func (l *ManagerOperator) managerActions(arrActions []dto.StateAction) {
	l.logger.Debug("manager actions", "trace", "docp-agent-os-instance.manager_operator.managerActions", "arrActions", arrActions)
	for _, act := range arrActions {
		switch act.Type {
		case "docp-agent":
			l.chanDocpAgent <- act
		case "datadog":
			l.chanDocpAgentDatadog <- act
		}
	}
	defer l.wg.Done()
}

func (l *ManagerOperator) verifyChangeMetadata(metadata []byte) bool {
	l.logger.Debug("verify change metadata", "trace", "docp-agent-os-instance.manager_operator.verifyChangeMetadata", "metadata", string(metadata))
	meta := libdto.Metadata{}
	if err := l.unmarshaller(metadata, &meta); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "verifyChangeMetadata", Priority: dto.ErrLevelMedium, Err: err}
		return true
	}
	computeInfoBytes, err := l.marshaller(meta.ComputeInfo)
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "verifyChangeMetadata", Priority: dto.ErrLevelMedium, Err: err}
		return true
	}
	hashMetadata := libutils.GenerateMd5Hash(computeInfoBytes)
	cacheHashMetadata := l.adapter.GetStore("metadata.hash")
	l.logger.Debug("verify change metadata", "trace", "docp-agent-os-instance.manager_operator.verifyChangeMetadata", "hashMetadata", hashMetadata, "cacheHashMetadata", cacheHashMetadata)
	if cacheHashMetadata == nil || hashMetadata != cacheHashMetadata.(string) {
		l.adapter.SetStore("metadata.hash", hashMetadata)
		return true
	}
	return false
}

// compareState execute compare for between received and current state
func (l *ManagerOperator) compareState() {
	l.logger.Debug("compare state", "trace", "docp-agent-os-instance.manager_operator.compareState")
	defer l.wg.Done()
	equals, err := l.adapter.CompareState()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "compareState", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	l.logger.Debug("compare state", "trace", "docp-agent-os-instance.manager_operator.compareState", "equals", equals)
	if equals && !l.validateIsComplete {
		l.validateIsComplete = true
	} else {
		// TODO: Implement logic for diference states
	}
	return
}

// validateState execute validation for state
func (l *ManagerOperator) validateState() {
	l.logger.Debug("validate state", "trace", "docp-agent-os-instance.manager_operator.validateState")
	defer l.wg.Done()
	if err := l.adapter.Validate(); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "validateState", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	return
}

// validateDocpAgentInstalled execute validation for docp agent if installed
func (l *ManagerOperator) validateDocpAgentInstalled() {
	l.logger.Debug("validate if docp agent is installed", "trace", "docp-agent-os-instance.manager_operator.validateDocpAgentInstalled")
	defer l.wg.Done()
	installed, err := l.adapter.ValidateDocpInstalled()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "validateDocpAgentInstalled", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	if installed {
	}
	notInstalled, err := l.adapter.ValidateDocpNotInstalled()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "validateDocpAgentInstalled", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	if notInstalled {
	}
}

// validateDatadogAgentInstalled execute validation for datadog agent if installed
func (l *ManagerOperator) validateDatadogAgentInstalled() {
	l.logger.Debug("validate if datadog agent is installed", "trace", "docp-agent-os-instance.manager_operator.validateDatadogAgentInstalled")
	defer l.wg.Done()
	installed, err := l.adapter.ValidateDatadogInstalled()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "validateDatadogAgentInstalled", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	if installed {
	}
	notInstalled, err := l.adapter.ValidateDatadogNotInstalled()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "validateDatadogAgentInstalled", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	if notInstalled {
	}
}

// validateDatadogAgentUpdateConfigs execute validation for datadog agent if update configs
func (l *ManagerOperator) validateDatadogAgentUpdateConfigs() {
	l.logger.Debug("validate if datadog agent is update configs", "trace", "docp-agent-os-instance.manager_operator.validateDatadogAgentUpdateConfigs")
	defer l.wg.Done()
	equals, err := l.adapter.CompareState()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "validateDatadogAgentUpdateConfigs", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	if equals {
	} else {
	}
}

// collectGetState collect state from service state check
func (l *ManagerOperator) collectGetState() {
	l.logger.Debug("collect get actions", "trace", "docp-agent-os-instance.manager_operator.collectGetActions")
	defer l.wg.Done()
	err := l.GetSignalFromStateCheck()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "collectGetState", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	return
}

// collectGetActions collect actions from service state check
func (l *ManagerOperator) collectGetActions() {
	l.logger.Debug("collect get actions", "trace", "docp-agent-os-instance.manager_operator.collectGetActions")
	defer l.wg.Done()
	var arrActions []dto.StateAction
	actions, err := l.GetActions()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "collectGetActions", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	if err := l.unmarshaller(actions, &arrActions); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "collectGetActions", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	l.wg.Add(1)
	go l.managerActions(arrActions)
}

// periodicGetActions execute periodic get actions the state
func (l *ManagerOperator) periodicTasks() {
	l.logger.Debug("periodic tasks", "trace", "docp-agent-os-instance.manager_operator.periodicTasks")
	defer l.wg.Done()

	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.wg.Add(3)
			go l.collectGetActions()
			go l.validateState()
			go l.compareState()
		case <-l.done:
			close(l.done)
			return
		}
	}
}

// periodicAutoUpdate execute periodic auto update
func (l *ManagerOperator) periodicAutoUpdate() {
	l.logger.Debug("periodic auto update", "trace", "docp-agent-os-instance.manager_operator.periodicAutoUpdate")
	defer l.wg.Done()

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.wg.Add(1)
			if err := l.AutoUpdateAgentVersion(); err != nil {
				l.logger.Error("error executing auto update agent version", "error", err.Error())
			}
			if err := l.AutoUpdateAgentDatadogVersion(); err != nil {
				l.logger.Error("error executing auto update agent datadog version", "error", err.Error())
			}
		case <-l.done:
			close(l.done)
			return
		}
	}
}

// periodicHandlerMetadata execute periodic handler metadata
func (l *ManagerOperator) periodicHandlerMetadata() {
	l.logger.Debug("periodic handler metadata", "trace", "docp-agent-os-instance.manager_operator.periodicHandlerMetadata")
	defer l.wg.Done()

	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.wg.Add(2)
			go l.getMetadata()
			go l.handleMetadata()
		case <-l.done:
			close(l.done)
			return
		}
	}
}

// persistLastSignalHashToStore execute save hash the received on store when restart/start agent
func (l *ManagerOperator) persistLastSignalHashToStore() error {
	receivedBytes, err := l.adapter.GetStateReceived()
	if err != nil {
		return err
	}

	receivedHash := libutils.GenerateMd5Hash(receivedBytes)
	if err := l.adapter.SetStore("signal.received", receivedHash); err != nil {
		return err
	}

	return nil
}

func (l *ManagerOperator) comunicateSCM() error {
	defer l.wg.Done()
	if err := l.adapter.HandlerSCMManager(); err != nil {
		return err
	}
	return nil
}
