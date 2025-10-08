package operators

import (
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/utils"

	adapters "github.com/OryaHub/agent-os-instance/libs/adapters"
	libinterfaces "github.com/OryaHub/agent-os-instance/libs/interfaces"
	libutils "github.com/OryaHub/agent-os-instance/libs/utils"
)

// UpdaterOperator is struct for updater the operator
type UpdaterOperator struct {
	logger          libinterfaces.ILogger
	adapter         *adapters.UpdaterAdapter
	delay           time.Duration
	maxRetry        int
	executeRollback bool
}

// NewUpdaterOperator return instance of updater operator
func NewUpdaterOperator() *UpdaterOperator {
	return &UpdaterOperator{
		delay:           time.Second * 1,
		maxRetry:        3,
		executeRollback: false,
	}
}

// Setup configure operator
func (l *UpdaterOperator) Setup() error {
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
	adapterUpdater := adapters.NewUpdaterAdapter(l.logger)
	if err := adapterUpdater.Prepare(); err != nil {
		return err
	}
	l.adapter = adapterUpdater
	return nil
}

// getVersionFromReceived extracts the version from the received content
func (l *UpdaterOperator) getVersionFromReceived() (string, error) {
	var applyVersion string
	received, err := l.adapter.GetContentReceived()
	if err != nil {
		return applyVersion, err
	}
	l.logger.Info("received content", "content", string(received))
	if len(received) > 0 {
		version, err := l.adapter.GetAgentVersionFromSignal(received)
		if err != nil {
			return applyVersion, err
		}
		l.logger.Info("version received", "version", version)
		if len(version) > 0 {
			if version == "latest" {
				l.logger.Info("latest version received, update needed")
				agentVersions, err := l.adapter.FetchAgentVersions()
				if err != nil {
					l.logger.Error("error fetching agent versions", "error", err.Error())
					return applyVersion, err
				}
				applyVersion = agentVersions.LatestVersion

			} else {
				applyVersion = version
			}

		}
	}
	if len(applyVersion) == 0 {
		return applyVersion, utils.ErrAgentVersionNotFound()
	}

	return applyVersion, nil
}

func (l *UpdaterOperator) RecoverServices() error {
	l.logger.Info("restarting services", "timestamp", time.Now())
	statusManager, err := l.adapter.Status("manager")
	if err != nil {
		return err
	}
	statusAgent, err := l.adapter.Status("agent")
	if err != nil {
		return err
	}

	if statusManager != "active" {
		if err := l.adapter.RestartService("manager"); err != nil {
			l.logger.Error("error restarting manager service", "error", err.Error())
			return err
		}
	}

	if statusAgent != "active" {
		if err := l.adapter.RestartService("agent"); err != nil {
			l.logger.Error("error restarting agent service", "error", err.Error())
			return err
		}
	}

	l.logger.Info("services restarted successfully", "timestamp", time.Now())
	return nil
}

// ExecuteUpdate execute update
func (l *UpdaterOperator) ExecuteUpdate() error {
	l.logger.Info("execute update", "timestamp", time.Now())
	applyVersion, err := l.getVersionFromReceived()
	if err != nil {
		return err
	}
	//apply version
	if err := l.adapter.ExecuteUpdateVersion(applyVersion); err != nil {
		return err
	}

	return nil
}

// Run execute loop the manager
func (l *UpdaterOperator) Run() error {
	if err := l.Setup(); err != nil {
		return err
	}
	l.logger.Info("execute updater running...")
	defer l.logger.Close()

	if err := l.Start(); err != nil {
		if errRecover := l.RecoverServices(); errRecover != nil {
			l.logger.Error("error recovering services", "error", errRecover.Error())

		}
		version, err := l.getVersionFromReceived()
		if err != nil {
			l.logger.Error("error getting version from received", "error", err.Error())
			return err
		}
		if err := l.adapter.UpdaterUninstall(version); err != nil {
			l.logger.Error("error uninstalling updater", "error", err.Error())
			return err
		}

	}

	return nil
}

// Start execute mathod for running in manager operator
func (l *UpdaterOperator) Start() error {
	l.logger.Info("execute start updater", "timestamp", time.Now())

	//execute update version
	if err := l.ExecuteUpdate(); err != nil {
		l.logger.Error("error executing update", "error", err.Error())
		return err
	}

	attempt := 1

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

loopvalidate:
	for {
		select {
		case <-ticker.C:
			success, err := l.adapter.ValidateSuccessUpdated()
			if err != nil {
				l.logger.Error("error validating update success", "error", err.Error())
				continue
			}
			l.logger.Info("validating update success", "success", success)
			if success {
				l.logger.Info("update successful", "timestamp", time.Now())
				break loopvalidate
			} else {
				attempt++
			}
			if attempt > l.maxRetry {
				l.logger.Info("maximum retry attempts reached", "maxRetry", l.maxRetry, "attempt", attempt)
				l.executeRollback = true
				break loopvalidate
			}

		}
	}

	//validate if need rollback
	if l.executeRollback {
		l.logger.Debug("executing roolback", "timestamp", time.Now())
		rollbackVersion, err := l.adapter.GetAgentRollbackVersion()
		if err != nil {
			l.logger.Error("error getting rollback version", "error", err.Error())
			return err
		}
		l.logger.Info("rollback version", "version", rollbackVersion)
		if err := l.adapter.ExecuteRollbackVersion(rollbackVersion); err != nil {
			l.logger.Error("error executing rollback version", "error", err.Error())
			return err
		}
	}
	version, err := l.getVersionFromReceived()
	if err != nil {
		l.logger.Error("error getting version from received", "error", err.Error())
		return err
	}

	//auto uninstall updater
	if err := l.adapter.UpdaterUninstall(version); err != nil {
		l.logger.Error("error uninstalling updater", "error", err.Error())
		return err
	}
	l.logger.Info("updater uninstalled successfully")
	return nil
}
