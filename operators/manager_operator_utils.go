package operators

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	libdto "github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/pkg"
	"github.com/OryaHub/agent-os-instance/libs/utils"
	libutils "github.com/OryaHub/agent-os-instance/libs/utils"
)

func (l *ManagerOperator) getLevelError() int {
	l.logger.Debug("get level error", "trace", "agent-os-instance.manager_operator.getLevelError")
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
	l.logger.Debug("get vendor latest version", "trace", "agent-os-instance.manager_operator.getVendorLatestVersion")
	latestVersion, err := l.vendorAdapter.GetLatestVersion()
	if err != nil {
		return "", err
	}
	return latestVersion, nil
}

// consumerErrors execute consumer for errors
func (l *ManagerOperator) consumerErrors() {
	l.logger.Debug("execute consumer errors", "trace", "agent-os-instance.manager_operator.consumerErrors")
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
	l.logger.Debug("resolve agent version", "trace", "agent-os-instance.manager_operator.resolveAgentVersion")
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
	l.logger.Debug("execute handle metadata", "trace", "agent-os-instance.linux_manager_operator.handleMetadata")
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
	l.logger.Debug("execute send metadata create", "trace", "agent-os-instance.linux_manager_operator.sendMetadataCreate")
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

			l.logger.Debug("execute send metadata create", "trace", "agent-os-instance.linux_manager_operator.sendMetadataCreate", "statusCode", statusCode)
			switch statusCode {
			case 202:
				if err := l.adapter.SaveInitialConfigFromRegister(result); err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataCreate", Priority: dto.ErrLevelMedium, Err: err}
					return
				}
			default:
				l.logger.Info("result from register service", "trace", "agent-os-instance.linux_manager_operator.sendMetadata", "result", string(result))
				if l.retryRegister <= l.maxRetry {
					go l.retryHandlerMetadata()
				}
			}
		}
	}
}

// sendMetadataUpdate execute send update metadata to register
func (l *ManagerOperator) sendMetadataUpdate() {
	l.logger.Debug("execute send metadata update", "trace", "agent-os-instance.linux_manager_operator.sendMetadataUpdate")
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

			l.logger.Debug("execute send metadata update", "trace", "agent-os-instance.linux_manager_operator.sendMetadataUpdate", "statusCode", statusCode)
			switch statusCode {
			case 202:
				configAgent, err := l.adapter.GetConfigAgent()
				if err != nil {
					l.chanErrors <- dto.CommonChanErrors{From: "sendMetadataUpdate", Priority: dto.ErrLevelMedium, Err: err}
				}

				var agentRegisterResponse libdto.AgentRegisterDataResponseSuccess
				if err := l.json.Unmarshall(result, &agentRegisterResponse); err != nil {
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
				l.logger.Info("result from register service", "trace", "agent-os-instance.linux_manager_operator.sendMetadata", "result", string(result))
			}
		}
	}
}

// validateDuplicatedSignal execute validate signal duplicated
// and notify transaction error with reason
func (l *ManagerOperator) validateDuplicatedSignal(signalBytes []byte) error {
	l.logger.Debug("validate duplicated signal", "trace", "agent-os-instance.manager_operator.validateDuplicatedSignal")
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
	l.logger.Debug("execute get metadata", "trace", "agent-os-instance.linux_manager_operator.getMetadata")
	for metadata := range l.adapter.Collect() {
		isChangedMetadata := l.verifyChangeMetadata(metadata)
		l.logger.Debug("execute get metadata", "trace", "agent-os-instance.linux_manager_operator.getMetadata", "isChangedMetadata", isChangedMetadata)
		if isChangedMetadata {
			l.chanMetadata <- metadata
		}
	}
	close(l.chanMetadata)
	l.wg.Done()
}

// extractDDApiKeyAndDDSiteFromEnvs return envs for install datadog agent
func (l *ManagerOperator) extractDDApiKeyAndDDSiteFromEnvs(envs []dto.StateActionEnvs) (string, string, error) {
	l.logger.Debug("extract envs the datadog", "trace", "agent-os-instance.manager_operator.extractDDApiKeyAndDDSiteFromEnvs", "envs", envs)
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
func (l *ManagerOperator) extractApmSingleStepEnvs(envs []dto.StateActionEnvs) (string, error) {
	l.logger.Debug("extract envs the datadog", "trace", "agent-os-instance.manager_operator.extractDDApiKeyAndDDSiteFromEnvs", "envs", envs)
	var ddApmInstrumentationLibraries string
	for _, env := range envs {
		if env.Name == "DD_APM_INSTRUMENTATION_LIBRARIES" {
			ddApmInstrumentationLibraries = env.Value
		}
	}

	if len(ddApmInstrumentationLibraries) == 0 {
		return "", errors.New("invalid dd apm instrumentation libraries")
	}
	return ddApmInstrumentationLibraries, nil
}

// extractApmTracingLibrayEnvs return envs for install datadog agent
func (l *ManagerOperator) extractApmTracingLibrayEnvs(envs []dto.StateActionEnvs) (string, string, string, error) {
	l.logger.Debug("extract envs the datadog", "trace", "agent-os-instance.manager_operator.extractApmTracingLibrayEnvs", "envs", envs)
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

// installAgent execute install the agent orya
func (l *ManagerOperator) installAgent() {
	l.logger.Debug("install the agent orya", "trace", "agent-os-instance.manager_operator.installAgent")
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

// installAgentDatadog execute call to api orya agent
// to install datadog agent
func (l *ManagerOperator) installAgentDatadog(ddApiKey, ddSite, version string) {
	l.logger.Debug("install agent datadog", "trace", "agent-os-instance.manager_operator.installAgentDatadog", "ddApiKey", ddApiKey, "ddSite", ddSite)
	defer l.wg.Done()
	status, err := l.adapter.Status("datadog")
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "installAgentDatadog", Priority: dto.ErrLevelHigh, Err: err}
		return
	}
	if status != "active" {
		result, err := l.adapter.OryaAgentApiInstallDatadog(ddApiKey, ddSite, version)
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
	l.logger.Debug("install and update agent datadog", "trace", "agent-os-instance.manager_operator.installAndUpdateAgentDatadog")
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
		flsBytes, err := l.json.Marshall(&fls)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return
		}

		l.wg.Add(1)
		go l.updateAgentDatadog(flsBytes)
	}
}

func (l *ManagerOperator) handlerInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationLibraries string) {
	l.logger.Debug("handle install datadog agent with APM single step", "trace", "agent-os-instance.manager_operator.handlerInstallDatadogWithApmSingleStep", "ddApiKey", ddApiKey, "ddSite", ddSite, "ddApmInstrumentationLibraries", ddApmInstrumentationLibraries)
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
				result, err := l.adapter.OryaAgentApiInstallDatadogWithApmSingleStep(
					ddApiKey, ddSite, ddApmInstrumentationLibraries,
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

// installAgentDatadog execute call to api orya agent
// to install datadog agent
func (l *ManagerOperator) installAgentDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationLibraries string) {
	l.logger.Debug("install agent datadog with apm single step", "trace", "agent-os-instance.manager_operator.installAgentDatadogWithApmSingleStep", "ddApiKey", ddApiKey, "ddSite", ddSite, "ddApmInstrumentationLibraries", ddApmInstrumentationLibraries)
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
	l.logger.Debug("install agent datadog with apm single step", "trace", "agent-os-instance.manager_operator.installAgentDatadogWithApmSingleStep", "oryaStatus", status, "alreadyTracer", alreadyTracer)
	if status != "active" {
		result, err := l.adapter.OryaAgentApiInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationLibraries)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "installAgentDatadogWithApmSingleStep", Priority: dto.ErrLevelMedium, Err: err}
			return
		}
		l.chanResultsApi <- result
		return
	} else if status == "active" && !alreadyTracer {
		l.wg.Add(1)
		go l.handlerInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationLibraries)
		return
	}
}

// installDatadogTracerWithTracingLibrary execute call to api orya agent
// to install datadog tracer with tracing library
func (l *ManagerOperator) installDatadogTracerWithTracingLibrary(ddApiKey, ddSite, language, pathTracer, version string) {
	l.logger.Debug("install datadog tracer", "trace", "agent-os-instance.manager_operator.installDatadogTracerWithTracingLibrary", "ddApiKey", ddApiKey, "ddSite", ddSite, "language", language, "pathTracer", pathTracer, "version", version)
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
	l.logger.Debug("install datadog tracer", "trace", "agent-os-instance.manager_operator.installDatadogTracerWithTracingLibrary", "oryaStatus", status, "alreadyTracer", alreadyTracer)
	existLanguage, err := l.adapter.ExistTracerLanguage(language)
	if err != nil {
		return
	}
	l.logger.Debug("install datadog tracer", "trace", "agent-os-instance.manager_operator.installDatadogTracerWithTracingLibrary", "language", language, "existLanguage", existLanguage)
	if status == "active" && !existLanguage {
		if err := l.adapter.AddTracerLanguage(language); err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "installDatadogTracerWithTracingLibrary", Priority: dto.ErrLevelMedium, Err: err}
			return
		}
		l.wg.Add(1)
		go l.adapter.OryaAgentApiInstallDatadogWithApmTracingLibrary(ddApiKey, ddSite, language, pathTracer, version)
		return
	}
}

// uninstallAgentDatadog exeuct call to api orya agent
// to uninstall datadog agent
func (l *ManagerOperator) uninstallAgentDatadog() {
	l.logger.Debug("uninstall agent datadog", "trace", "agent-os-instance.manager_operator.uninstallAgentDatadog")
	defer l.wg.Done()

	result, err := l.adapter.OryaAgentApiUninstallDatadog()
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

// updateAgentDatadog execute call to api orya agent
// to update datadog agent
func (l *ManagerOperator) updateAgentDatadog(content []byte) {
	l.logger.Debug("update agent datadog", "trace", "agent-os-instance.manager_operator.updateAgentDatadog", "content", string(content))
	defer l.wg.Done()

	newHash := utils.GenerateMd5Hash(content)
	existsDatadogHash := l.adapter.GetStore("update.datadog.hash")

	if existsDatadogHash == nil {
		result, err := l.adapter.OryaAgentApiUpdateConfigurationsDatadog(content)
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
	l.logger.Debug("auto uninstall with other vendors the manager", "trace", "agent-os-instance.manager_operator.autoUninstallWithOtherVendors")
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

	// execute autouninstall for orya agent
	transaction := libutils.NewTransactionStatus()
	ctx := context.WithValue(context.Background(), libdto.ContextTransactionStatus, transaction)

	go l.adapter.NotifyStatus("uninstall_orya_received", pkg.TransactionEventOpen, "uninstall orya received", ctx)
	time.Sleep(l.delay)

	go l.adapter.NotifyStatus("uninstall_orya_processing", pkg.TransactionEventUpdate, "uninstall orya processing", ctx)
	time.Sleep(l.delay)

	//get version installed
	version, err := l.resolveAgentVersion()
	if err != nil {
		go l.adapter.NotifyStatus("uninstall_orya_error", pkg.TransactionEventClose, "failed uninstall orya", ctx)
		l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
		return
	}

	if err := l.adapter.AutoUninstall(version); err != nil {
		go l.adapter.NotifyStatus("uninstall_orya_error", pkg.TransactionEventClose, "failed uninstall orya", ctx)
		l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
		return
	}
	go l.adapter.NotifyStatus("uninstall_orya_completed", pkg.TransactionEventClose, "uninstall orya completed", ctx)

	return
}

// autoUninstallAgent execute auto uninstall the manager
func (l *ManagerOperator) autoUninstall() {
	l.logger.Debug("auto uninstall the manager", "trace", "agent-os-instance.manager_operator.autoUninstall")

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

		go l.adapter.NotifyStatus("uninstall_orya_received", pkg.TransactionEventOpen, "uninstall orya received", ctx)
		time.Sleep(l.delay)

		defer l.wg.Done()

		go l.adapter.NotifyStatus("uninstall_orya_processing", pkg.TransactionEventUpdate, "uninstall orya processing", ctx)
		time.Sleep(l.delay)

		//get version
		version, err := l.resolveAgentVersion()
		if err != nil {
			go l.adapter.NotifyStatus("uninstall_orya_error", pkg.TransactionEventClose, "failed uninstall orya", ctx)
			l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
			return
		}

		if err := l.adapter.AutoUninstall(version); err != nil {
			go l.adapter.NotifyStatus("uninstall_orya_error", pkg.TransactionEventClose, "failed uninstall orya", ctx)
			l.chanErrors <- dto.CommonChanErrors{From: "autoUninstall", Priority: dto.ErrLevelHigh, Err: err}
			return
		}

		go l.adapter.NotifyStatus("uninstall_orya_completed", pkg.TransactionEventClose, "uninstall orya completed", ctx)
	}
	return
}

// consumerResultsFromApiOryaAgent execute consume the result
// from api orya agent
func (l *ManagerOperator) consumerResultsFromApiOryaAgent() {
	l.logger.Debug("consumer results from api orya agent", "trace", "agent-os-instance.manager_operator.consumerResultsFromApiOryaAgent")
	defer l.wg.Done()
	for res := range l.chanResultsApi {
		l.logger.Debug("consumer results from api orya agent", "trace", "agent-os-instance.manager_operator.consumerResultsFromApiOryaAgent", "result", string(res))
	}
}

// consumeActionsOryaAgent execute consume the actions the agent
func (l *ManagerOperator) consumeActionsOryaAgent() {
	l.logger.Debug("consume actions orya agent", "trace", "agent-os-instance.manager_operator.consumeActionsOryaAgent")
	defer l.wg.Done()
	for act := range l.chanOryaAgent {
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
	l.logger.Debug("consumer actions datadog", "trace", "agent-os-instance.manager_operator.consumerActionsDatadog")
	defer l.wg.Done()

	// get actions for datadog
	for act := range l.chanOryaAgentDatadog {
		// verify if datadog already installed
		datadogAlreadyInstalled, err := l.adapter.AlreadyInstalled("datadog")
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
			return
		}
		l.logger.Debug("consumer actions datadog", "trace", "agent-os-instance.manager_operator.consumerActionsDatadog", "action", act)
		// action update configurations datadog
		if act.Action == "update" {
			for _, fls := range act.Files {
				flsBytes, err := l.json.Marshall(&fls)
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

		l.logger.Debug("consumer actions datadog", "trace", "agent-os-instance.manager_operator.consumerActionsDatadog", "datadogAlreadyInstalled", datadogAlreadyInstalled)
		if err != nil {
			l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
		}
		if act.Action == "install" {
			version := act.Version
			ddApiKey, ddSite, err := l.extractDDApiKeyAndDDSiteFromEnvs(act.Envs)
			if err != nil {
				l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
				return
			}
			if act.Component == "tracer" {
				if act.Mode == "single_step" {
					if datadogAlreadyInstalled {
						ddApmInstrumentationLibraries, err := l.extractApmSingleStepEnvs(act.ComponentEnvs)
						if err != nil {
							l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
							return
						}
						l.wg.Add(1)
						go l.handlerInstallDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationLibraries)
					} else {
						l.wg.Add(1)
						ddApmInstrumentationLibraries, err := l.extractApmSingleStepEnvs(act.ComponentEnvs)
						if err != nil {
							l.chanErrors <- dto.CommonChanErrors{From: "consumerActionsDatadog", Priority: dto.ErrLevelMedium, Err: err}
							return
						}
						go l.installAgentDatadogWithApmSingleStep(ddApiKey, ddSite, ddApmInstrumentationLibraries)
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
						go l.installAgentDatadog(ddApiKey, ddSite, version)
						go l.handlerUpdateAgentDatadogAfterInstall(act.Files)
					} else {
						l.wg.Add(1)
						go l.installAgentDatadog(ddApiKey, ddSite, version)
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
	l.logger.Debug("consume all actions", "trace", "agent-os-instance.manager_operator.consumeAllActions")
	defer l.wg.Done()
	l.wg.Add(2)
	go l.consumeActionsOryaAgent()
	go l.consumerActionsDatadog()
}

// managerActions execut segment actions by type
func (l *ManagerOperator) managerActions(arrActions []dto.StateAction) {
	l.logger.Debug("manager actions", "trace", "agent-os-instance.manager_operator.managerActions", "arrActions", arrActions)
	for _, act := range arrActions {
		switch act.Type {
		case "orya-agent":
			l.chanOryaAgent <- act
		case "datadog":
			l.chanOryaAgentDatadog <- act
		}
	}
	defer l.wg.Done()
}

func (l *ManagerOperator) verifyChangeMetadata(metadata []byte) bool {
	l.logger.Debug("verify change metadata", "trace", "agent-os-instance.manager_operator.verifyChangeMetadata", "metadata", string(metadata))
	meta := libdto.Metadata{}
	if err := l.json.Unmarshall(metadata, &meta); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "verifyChangeMetadata", Priority: dto.ErrLevelMedium, Err: err}
		return true
	}
	computeInfoBytes, err := l.json.Marshall(meta.ComputeInfo)
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "verifyChangeMetadata", Priority: dto.ErrLevelMedium, Err: err}
		return true
	}
	hashMetadata := libutils.GenerateMd5Hash(computeInfoBytes)
	cacheHashMetadata := l.adapter.GetStore("metadata.hash")
	l.logger.Debug("verify change metadata", "trace", "agent-os-instance.manager_operator.verifyChangeMetadata", "hashMetadata", hashMetadata, "cacheHashMetadata", cacheHashMetadata)
	if cacheHashMetadata == nil || hashMetadata != cacheHashMetadata.(string) {
		l.adapter.SetStore("metadata.hash", hashMetadata)
		return true
	}
	return false
}

// compareState execute compare for between received and current state
func (l *ManagerOperator) compareState() {
	l.logger.Debug("compare state", "trace", "agent-os-instance.manager_operator.compareState")
	defer l.wg.Done()
	equals, err := l.adapter.CompareState()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "compareState", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	l.logger.Debug("compare state", "trace", "agent-os-instance.manager_operator.compareState", "equals", equals)
	if equals && !l.validateIsComplete {
		l.validateIsComplete = true
	} else {
		// TODO: Implement logic for diference states
	}
	return
}

// validateState execute validation for state
func (l *ManagerOperator) validateState() {
	l.logger.Debug("validate state", "trace", "agent-os-instance.manager_operator.validateState")
	defer l.wg.Done()
	if err := l.adapter.Validate(); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "validateState", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	return
}

// collectGetState collect state from service state check
func (l *ManagerOperator) collectGetState() {
	l.logger.Debug("collect get actions", "trace", "agent-os-instance.manager_operator.collectGetActions")
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
	l.logger.Debug("collect get actions", "trace", "agent-os-instance.manager_operator.collectGetActions")
	defer l.wg.Done()
	var arrActions []dto.StateAction
	actions, err := l.GetActions()
	if err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "collectGetActions", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	if err := l.json.Unmarshall(actions, &arrActions); err != nil {
		l.chanErrors <- dto.CommonChanErrors{From: "collectGetActions", Priority: dto.ErrLevelMedium, Err: err}
		return
	}
	l.wg.Add(1)
	go l.managerActions(arrActions)
}

// periodicGetActions execute periodic get actions the state
func (l *ManagerOperator) periodicTasks() {
	l.logger.Debug("periodic tasks", "trace", "agent-os-instance.manager_operator.periodicTasks")
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
	l.logger.Debug("periodic auto update", "trace", "agent-os-instance.manager_operator.periodicAutoUpdate")
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
	l.logger.Debug("periodic handler metadata", "trace", "agent-os-instance.manager_operator.periodicHandlerMetadata")
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
