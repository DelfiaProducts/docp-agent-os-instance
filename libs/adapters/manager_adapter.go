package adapters

import (
	"net/http"
	"sync"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/components"
	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/interfaces"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/services"
	"github.com/OryaHub/agent-os-instance/libs/utils"
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
	json                     *pkg.JsonClient
	client                   *http.Client
	agentWorkDir             string
	store                    *utils.Store
	stateCheck               *services.StateCheckService
	auth                     *services.AuthService
	utilityService           *services.UtilityService
	oryaApiPort              string
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
		json:                     pkg.NewJsonClient(),
		ymlClient:                pkg.NewYmlClient(),
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
	chanLinuxMetrics := make(chan []byte, 1)
	chanClose := make(chan struct{}, 1)
	wg := &sync.WaitGroup{}
	hostStats := pkg.NewHostStats()
	l.interval = interval
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
	l.oryaApiPort = apiPort
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

// DaemonReload execute daemon reload the service in systemd
func (l *ManagerAdapter) DaemonReload() error {
	l.logger.Info("daemon reload", "timestamp", time.Now().UTC())
	l.logger.Debug("daemon reload", "trace", "agent-os-instance.manager_adapter.DaemonReload")
	if err := l.osOperation.DaemonReload(); err != nil {
		return err
	}
	return nil
}

// RestartService execute restart the service in systemd
func (l *ManagerAdapter) RestartService(serviceName string) error {
	l.logger.Info("restart service", "timestamp", time.Now().UTC())
	l.logger.Debug("restart service", "trace", "agent-os-instance.manager_adapter.RestartService")
	if err := l.osOperation.RestartService(serviceName); err != nil {
		return err
	}
	return nil
}

// StopService execute stop the service in systemd
func (l *ManagerAdapter) StopService(serviceName string) error {
	l.logger.Info("stop service", "timestamp", time.Now().UTC())
	l.logger.Debug("stop service", "trace", "agent-os-instance.manager_adapter.StopService")
	if err := l.osOperation.StopService(serviceName); err != nil {
		return err
	}
	return nil
}

// InstallAgent execute install the agent orya
func (l *ManagerAdapter) InstallAgent(version string) error {
	l.logger.Info("install agent", "timestamp", time.Now().UTC(), "version", version)
	l.logger.Debug("install agent", "trace", "agent-os-instance.manager_adapter.InstallAgent")
	if err := l.osOperation.InstallAgent(version); err != nil {
		return err
	}
	return nil
}

// UninstallAgent execute uninstall the agent orya
func (l *ManagerAdapter) UninstallAgent(version string) error {
	l.logger.Info("uninstall agent", "timestamp", time.Now().UTC(), "version", version)
	l.logger.Debug("uninstall agent", "trace", "agent-os-instance.manager_adapter.UninstallAgent")
	if err := l.osOperation.UninstallAgent(version); err != nil {
		return err
	}
	return nil
}

// UpdateAgent execute update the agent orya
func (l *ManagerAdapter) UpdateAgent(version string) error {
	l.logger.Info("update agent", "timestamp", time.Now().UTC(), "version", version)
	l.logger.Debug("update agent", "trace", "agent-os-instance.manager_adapter.UpdateAgent")
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
func (l *ManagerAdapter) AutoUninstall(version string) error {
	l.logger.Info("auto uninstall agent", "timestamp", time.Now().UTC(), "version", version)
	l.logger.Debug("auto uninstall agent", "trace", "agent-os-instance.manager_adapter.AutoUninstall")
	if err := l.osOperation.AutoUninstall(version); err != nil {
		return err
	}
	return nil
}
