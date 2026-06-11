package operators

import (
	"context"
	"encoding/json"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"

	adapters "github.com/OryaHub/agent-os-instance/libs/adapters"
	libdto "github.com/OryaHub/agent-os-instance/libs/dto"
	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	services "github.com/OryaHub/agent-os-instance/libs/services"
	"github.com/OryaHub/agent-os-instance/libs/utils"
	libutils "github.com/OryaHub/agent-os-instance/libs/utils"
)

// ManagerOperator is struct for manager the operator
type ManagerOperator struct {
	filePath             string
	logger               libinterfaces.ILogger
	adapter              *adapters.ManagerAdapter
	vendorAdapter        *adapters.DatadogAdapter
	register             *services.AgentRegisterService
	stateCheck           *services.StateCheckService
	json                 *pkg.JsonClient
	wg                   *sync.WaitGroup
	done                 chan struct{}
	chanErrors           chan dto.CommonChanErrors
	chanMetadata         chan []byte
	chanResultsApi       chan []byte
	chanOryaAgent        chan dto.StateAction
	chanOryaAgentDatadog chan dto.StateAction
	validateIsComplete   bool
	retryRegister        int
	maxRetry             int
	delay                time.Duration
	intervalGetSignal    time.Duration
	tickerSignal         *time.Ticker
}

// NewManagerOperator return instance of manager operator
func NewManagerOperator() *ManagerOperator {
	return &ManagerOperator{
		wg:                   &sync.WaitGroup{},
		json:                 pkg.NewJsonClient(),
		done:                 make(chan struct{}),
		chanErrors:           make(chan dto.CommonChanErrors, 1),
		chanMetadata:         make(chan []byte, 1),
		chanResultsApi:       make(chan []byte, 1),
		chanOryaAgent:        make(chan dto.StateAction, 1),
		chanOryaAgentDatadog: make(chan dto.StateAction, 1),
		validateIsComplete:   false,
		retryRegister:        0,
		maxRetry:             10,
		delay:                time.Second * 1,
		intervalGetSignal:    time.Minute * 1,
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
		loggerFile := libutils.NewOryaLoggerWindowsFileText(logPath)
		logger = loggerFile
	} else {
		logger = libutils.NewOryaLoggerJSON(os.Stdout)
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

	ticker := time.NewTicker(l.intervalGetSignal)
	l.tickerSignal = ticker
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
	l.logger.Info("get signal", "timestamp", time.Now().UTC())
	l.logger.Debug("get signal from state check", "trace", "agent-os-instance.manager_operator.GetSignalFromStateCheck")
	stateCheckBytes, statusCode, err := l.stateCheck.GetState()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "GetState", Priority: dto.ErrLevelMedium, Err: err}
	}
	l.logger.Debug("get state", "trace", "agent-os-instance.manager_operator.GetState", "statusCode", statusCode)
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
		l.logger.Debug("get signal from state check", "trace", "agent-os-instance.manager_operator.GetSignalFromStateCheck", "status", "not state present")
		return nil
	case 403:
		l.logger.Debug("get signal from state check", "trace", "agent-os-instance.manager_operator.GetSignalFromStateCheck", "status", "not authorized")
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

// UpdateAgent execute update the agent orya
func (l *ManagerOperator) UpdateAgent(version string) error {
	l.logger.Info("update the agent", "timestamp", time.Now().UTC(), "version", version)
	l.logger.Debug("update the agent orya", "trace", "agent-os-instance.manager_operator.updateAgent")
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

			go l.adapter.NotifyStatus("update_orya_received", pkg.TransactionEventOpen, "update orya received", ctx)
			time.Sleep(l.delay)

			go l.adapter.NotifyStatus("update_orya_initiate", pkg.TransactionEventUpdate, "update orya initialized", ctx)
			time.Sleep(l.delay)

			if err := l.adapter.UpdateAgent(version); err != nil {
				go l.adapter.NotifyStatus("update_orya_error", pkg.TransactionEventClose, "failed update agent version", ctx)
				l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
				return err
			}

			//save new version
			if err := l.adapter.SaveAgentVersion(version); err != nil {
				go l.adapter.NotifyStatus("update_orya_error", pkg.TransactionEventClose, "failed save agent version", ctx)
				l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
				return err
			}

			//save rollback version
			if err := l.adapter.SaveAgentRollbackVersion(agentVersion); err != nil {
				go l.adapter.NotifyStatus("update_orya_error", pkg.TransactionEventClose, "failed save agent version", ctx)
				l.chanErrors <- dto.CommonChanErrors{From: "updateAgent", Priority: dto.ErrLevelHigh, Err: err}
				return err
			}

			go l.adapter.NotifyStatus("update_orya_completed", pkg.TransactionEventClose, "update orya completed", ctx)
			return nil
		}
	}
	return nil
}

// GetActions get action for execute in manager
func (l *ManagerOperator) GetActions() ([]byte, error) {
	l.logger.Debug("get actions", "trace", "agent-os-instance.manager_operator.GetActions")
	actions, err := l.adapter.GetState()
	l.logger.Debug("get actions", "trace", "agent-os-instance.manager_operator.GetActions", "actions", string(actions))
	if err != nil {
		return nil, err
	}
	return actions, nil
}

// Stop execute stoping the ManagerOperator
func (l *ManagerOperator) Stop() {
	l.logger.Debug("stop", "trace", "agent-os-instance.manager_operator.Stop")
	l.done <- struct{}{}
}

// AutoUpdateAgentVersion execute auto update agent version
func (l *ManagerOperator) AutoUpdateAgentVersion() error {
	received, err := l.adapter.GetStateReceived()
	if err != nil {
		return err
	}
	l.logger.Info("auto update agent version", "received", string(received))

	var stateCheck dto.StateCheckResponse
	if err := l.json.Unmarshall(received, &stateCheck); err != nil {
		return err
	}
	if !stateCheck.Signal.Agents.OryaAgent.AutoUpdate {
		l.logger.Info("auto update agent version disabled by auto_update=false, skipping")
		return nil
	}

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
	l.logger.Info("update agent version vendor", "timestamp", time.Now().UTC(), "version", version)
	l.logger.Debug("update agent version datadog", "trace", "agent-os-instance.manager_operator.updateAgentVersionDatadog")
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
	_, err = l.adapter.OryaAgentApiUpdateVersionDatadog(version)
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
		l.logger.Debug("rollback update vendor version", "trace", "agent-os-instance.manager_operator.UpdateAgentVersionDatadog")
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
	l.logger.Info("auto update agent datadog version", "timestamp", time.Now().UTC())
	received, err := l.adapter.GetStateReceived()
	if err != nil {
		return err
	}
	l.logger.Info("auto update agent datadog version", "received", string(received))

	var stateCheck dto.StateCheckResponse
	if err := l.json.Unmarshall(received, &stateCheck); err != nil {
		return err
	}
	if !stateCheck.Signal.Agents.DatadogAgent.AutoUpdate {
		l.logger.Info("auto update agent datadog version disabled by auto_update=false, skipping")
		return nil
	}

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
	l.logger.Debug("execute running", "trace", "agent-os-instance.manager_operator.Run")
	defer l.logger.Close()
	l.wg.Add(14)
	go l.persistLastSignalHashToStore()
	go l.comunicateSCM()
	go l.Profiling()
	go l.Start()
	go l.consumerErrors()
	go l.consumerResultsFromApiOryaAgent()
	go l.collectGetState()
	go l.collectGetActions()
	go l.consumeAllActions()
	go l.handleMetadata()
	go l.periodicHandlerMetadata()
	go l.periodicAutoUpdate()
	go l.periodicTasks()
	go l.getMetadata()

	// On restart, the Datadog agent may not yet be active when the initial
	// metadata is collected, leaving vendors_info.datadog empty. Wait for it
	// to become active and then trigger a metadata update so the register
	// service receives the vendor info (version, host ID, etc.).
	if installed, err := l.adapter.AlreadyInstalled("datadog"); err == nil && installed {
		l.wg.Add(1)
		go l.waitAgentAndSendMetadata()
	}

	l.wg.Wait()
	return nil
}

// Start execute mathod for running in manager operator
func (l *ManagerOperator) Start() {
	l.logger.Debug("start tasks", "trace", "agent-os-instance.manager_operator.Start")
	defer l.wg.Done()

	l.wg.Add(1)
	go l.installAgent()

	defer l.tickerSignal.Stop()

	for {
		select {
		case <-l.tickerSignal.C:
			l.wg.Add(1)
			go l.collectGetState()

		case <-l.done:
			close(l.done)
			return
		}
	}
}

func (l *ManagerOperator) CheckHealth() {
	l.logger.Debug("in CheckHealth", "trace", "agent-os-instance.manager_operator.CheckHealth")
	resp, err := http.Get("http://localhost:3000/health")
	if err != nil {
		l.logger.Error("in CheckHealth", "trace", "agent-os-instance.manager_operator.CheckHealth", "error", err.Error())
		return
	}
	defer resp.Body.Close()

	var healthResp dto.HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&healthResp)
	if err != nil {
		l.logger.Error("in CheckHealth", "trace", "agent-os-instance.manager_operator.CheckHealth", "error", err.Error())
		return
	}

	if healthResp.Status == "success" && healthResp.Code == "HEALTH_OK" {
		l.logger.Debug("check health success", "trace", "agent-os-instance.manager_operator.CheckHealth", "healthResp", healthResp)
	} else {
		l.logger.Warn("check health failed", "trace", "agent-os-instance.manager_operator.CheckHealth", "healthResp", healthResp)
	}
}

func (l *ManagerOperator) Profiling() {
	l.logger.Debug("profiling task", "trace", "agent-os-instance.manager_operator.Profiling")
	defer l.wg.Done()
	if err := http.ListenAndServe(":4040", nil); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "Profiling", Priority: dto.ErrLevelLow, Err: err}
	}
}

// Close execute close the operator loop
func (l *ManagerOperator) Close() error {
	l.logger.Debug("execute close", "trace", "agent-os-instance.linux_manager_operator.Close")
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		l.adapter.Close()
	}()
	return nil
}
