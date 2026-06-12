package adapters

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// getInfos execute get the infos in host
func (l *ManagerAdapter) getInfos() {
	l.logger.Debug("get infos", "trace", "agent-os-instance.manager_adapter.getInfos")
	computeInfo, err := l.hostStats.ComputeInfo()
	if err != nil {
		l.logger.Error("error host info", "trace", "agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	cpuInfo, err := l.hostStats.CPUInfo()
	if err != nil {
		l.logger.Error("error cpu info", "trace", "agent-os-instance.linux_adapter.getInfos", "error", err.Error())
	}
	memoryInfo, err := l.hostStats.MemoryInfo()
	if err != nil {
		l.logger.Error("error memory info", "trace", "agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	diskInfo, err := l.hostStats.DiskInfo()
	if err != nil {
		l.logger.Error("error disk info", "trace", "agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	processInfo, err := l.hostStats.ProcessInfo()
	if err != nil {
		l.logger.Error("error process info", "trace", "agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	linuxMetadata := dto.Metadata{
		ComputeInfo:  computeInfo,
		CPUInfo:      cpuInfo,
		MemoryInfo:   memoryInfo,
		DiskInfo:     diskInfo,
		ProcessInfos: processInfo,
	}
	datadogStatus, err := l.osOperation.Status("datadog")
	if err != nil {
		l.logger.Error("error datadog status", "trace", "agent-os-instance.manager_adapter.getInfos", "error", err.Error())
	}
	if datadogStatus == "active" {
		infosBytes, err := l.vendorOperation.GetInfos()
		if err != nil {
			l.logger.Error("error get datadog infos", "trace", "agent-os-instance.manager_adapter.getInfos", "error", err.Error())
		}
		var datadogInfos dto.DatadogInfos
		if err := l.json.Unmarshall(infosBytes, &datadogInfos); err != nil {
			l.logger.Error("error unmarshal datadog infos", "trace", "agent-os-instance.manager_adapter.getInfos", "error", err.Error())
		} else {
			linuxMetadata.VendorsInfo.Datadog = datadogInfos
		}
	}

	if !l.isClosed {
		metadataBytes, err := l.json.Marshall(linuxMetadata)
		if err != nil {
			l.logger.Error("error in marshal metadata", "trace", "agent-os-instance.manager_adapter.getInfos", "error", err.Error())
		}
		l.chanMetadata <- metadataBytes
	}
	l.wg.Done()
}

// firstGetInfos execute first get infos in host
func (l *ManagerAdapter) firstGetInfos() {
	l.logger.Debug("first get infos", "trace", "agent-os-instance.manager_adapter.firstGetInfos")
	l.wg.Add(1)
	time.Sleep(time.Second * 5)
	go l.getInfos()
	l.wg.Done()
}

// closeChannels closing channels
func (l *ManagerAdapter) closeChannels() {
	l.logger.Debug("close channels", "trace", "agent-os-instance.manager_adapter.closeChannels")
	close(l.chanClose)
	close(l.chanMetadata)
	l.isClosed = true
	l.wg.Done()
}

// requestForAgentInstallDatadog execute request for agent
func (l *ManagerAdapter) requestForAgentInstallDatadog(url string, method string, data []byte) ([]byte, error) {
	l.logger.Debug("request for agent install datadog", "trace", "agent-os-instance.manager_adapter.requestForAgentInstallDatadog", "url", url, "method", method, "data", data)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		l.logger.Error("error in create request", "trace", "agent-os-instance.manager_adapter.requestForAgentInstallDatadog", "error", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := l.client.Do(req)
	if err != nil {
		l.logger.Error("error in execute request", "trace", "agent-os-instance.manager_adapter.requestForAgentInstallDatadog", "error", err.Error())
		return nil, err
	}
	defer res.Body.Close()
	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		l.logger.Error("error in read body response", "trace", "agent-os-instance.manager_adapter.requestForAgentInstallDatadog", "error", err.Error())
		return nil, err
	}
	return respBytes, nil
}

// prepareOryaAgentAction execute prepare for orya agent action
func (l *ManagerAdapter) prepareOryaAgentAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare orya agent action", "trace", "agent-os-instance.manager_adapter.prepareOryaAgentAction", "stateCheckSignal", stateCheckSignal)
	if stateCheckSignal.TypeSignal == "update" {
		if len(stateCheckSignal.Agents.OryaAgent.Version) > 0 {
			return dto.StateAction{
				Type:    "orya-agent",
				Action:  "update",
				Version: stateCheckSignal.Agents.OryaAgent.Version,
			}
		}
	} else if stateCheckSignal.TypeSignal == "uninstall" {
		return dto.StateAction{
			Type:   "orya-agent",
			Action: "uninstall",
		}
	}
	return dto.StateAction{}
}

func (l *ManagerAdapter) prepareOryaStandbyAgentAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare orya standby agent action", "trace", "agent-os-instance.manager_adapter.prepareOryaStandbyAgentAction", "stateCheckSignal", stateCheckSignal)
	if stateCheckSignal.TypeSignal == "standby" {
		if len(stateCheckSignal.Mode) > 0 {
			standbyAction := dto.StateAction{
				Type:   "orya-agent",
				Action: "standby",
			}
			mode := stateCheckSignal.Mode
			switch mode {
			case "start":
				standbyAction.Sleep = stateCheckSignal.Sleep
			case "stop":
				standbyAction.Sleep = 1
			}
			return standbyAction

		}
	}
	return dto.StateAction{}
}

// getActionsFiles return slice the files from configurations
func (l *ManagerAdapter) getActionsFiles(configurations dto.StateCheckDatadogConfigurations) []dto.StateActionFiles {
	l.logger.Debug("get actions files", "trace", "agent-os-instance.manager_adapter.getActionsFiles", "configurations", configurations)
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

// extractHostnameFromFiles looks through the files slice for one whose
// FilePath references a datadog.yaml, then extracts the hostname value
// from its YAML content. Returns empty string if not found.
func (l *ManagerAdapter) extractHostnameFromFiles(files []dto.StateActionFiles) string {
	for _, f := range files {
		if strings.HasSuffix(f.FilePath, "datadog.yaml") || strings.HasSuffix(f.FilePath, "datadog.yml") {
			hostname := utils.ExtractHostnameFromDatadogConfig(f.Content)
			if hostname != "" {
				l.logger.Debug("extracted hostname from datadog config file", "trace", "agent-os-instance.manager_adapter.extractHostnameFromFiles", "hostname", hostname)
				return hostname
			}
		}
	}
	return ""
}

// parseSiteDatadog execute parse the site datadog
func (l *ManagerAdapter) parseSiteDatadog(site string) string {
	l.logger.Debug("parse site datadog", "trace", "agent-os-instance.manager_adapter.parseSiteDatadog", "site", site)
	defaultSite := "datadoghq.com"
	host, err := utils.GetBaseUrlSite(site)
	if err != nil {
		return defaultSite
	}
	splited := strings.Split(host, ".")
	if len(splited) > 2 || len(host) == 0 {
		return defaultSite
	}

	return host
}

// prepareTracerDatadogSingleStepAction execute prepare for tracer datadog single step action
func (l *ManagerAdapter) prepareTracerDatadogSingleStepAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare tracer datadog library action", "trace", "agent-os-instance.manager_adapter.prepareTracerDatadogSingleStepAction", "stateCheckSignal", stateCheckSignal)
	var action dto.StateAction
	var envVars []dto.StateActionEnvs
	if stateCheckSignal.TypeSignal == "update" {
		if stateCheckSignal.Agents.DatadogAgent.Enabled {
			if len(stateCheckSignal.Agents.DatadogTracerSingleStep.DDApmInstrumentationLibraries) > 0 && len(stateCheckSignal.Agents.DatadogTracerLibrary.Version) == 0 {
				tracerSingleStep := stateCheckSignal.Agents.DatadogTracerSingleStep
				var version string
				var files []dto.StateActionFiles
				var componetEnvVars []dto.StateActionEnvs

				componetEnvVars = append(componetEnvVars, dto.StateActionEnvs{
					Name:  "DD_APM_INSTRUMENTATION_LIBRARIES",
					Value: tracerSingleStep.DDApmInstrumentationLibraries,
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
					version = datadogAgent.Version
				}

				action = dto.StateAction{
					Type:          "datadog",
					Action:        "install",
					Mode:          "single_step",
					Component:     "tracer",
					Version:       version,
					HostTags:      stateCheckSignal.HostTags,
					ComponentEnvs: componetEnvVars,
					Envs:          envVars,
					Files:         files,
				}
			}
		}

	}
	return action
}

// prepareTracerDatadogLibraryAction execute prepare for tracer datadog tracer library action
func (l *ManagerAdapter) prepareTracerDatadogLibraryAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare tracer datadog library action", "trace", "agent-os-instance.manager_adapter.prepareTracerDatadogLibraryAction", "stateCheckSignal", stateCheckSignal)
	var action dto.StateAction
	var envVars []dto.StateActionEnvs
	if stateCheckSignal.TypeSignal == "update" {
		if stateCheckSignal.Agents.DatadogAgent.Enabled {
			if len(stateCheckSignal.Agents.DatadogTracerLibrary.Version) > 0 && len(stateCheckSignal.Agents.DatadogTracerSingleStep.DDApmInstrumentationLibraries) == 0 {
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
					HostTags:      stateCheckSignal.HostTags,
					Envs:          envVars,
					Files:         files,
				}
			}
		}

	}
	return action
}

// prepareAgentDatadogAction execute prepare for agent datadog action
func (l *ManagerAdapter) prepareAgentDatadogAction(stateCheckSignal dto.StateCheckSignal) dto.StateAction {
	l.logger.Debug("prepare agent datadog action", "trace", "agent-os-instance.manager_adapter.prepareAgentDatadogAction", "stateCheckSignal", stateCheckSignal)
	var action dto.StateAction
	var envVars []dto.StateActionEnvs
	var componetEnvVars []dto.StateActionEnvs
	var files []dto.StateActionFiles
	var version string

	if stateCheckSignal.TypeSignal == "update" {
		if stateCheckSignal.Agents.DatadogAgent.Enabled {
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
				version = datadogAgent.Version
				l.logger.Debug("prepare agent datadog action", "files", files)
				return dto.StateAction{
					Type:          "datadog",
					Action:        "install",
					Mode:          "",
					Component:     "agent",
					ComponentEnvs: componetEnvVars,
					Envs:          envVars,
					Files:         files,
					Version:       version,
					HostTags:      stateCheckSignal.HostTags,
					Hostname:      l.extractHostnameFromFiles(files),
				}
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
	l.logger.Debug("prepare agent datadog action", "trace", "agent-os-instance.manager_adapter.prepareAgentDatadogAction", "stateCheckSignal", stateCheckSignal)
	var action dto.StateAction
	var envVars []dto.StateActionEnvs
	var componetEnvVars []dto.StateActionEnvs
	var files []dto.StateActionFiles

	if stateCheckSignal.TypeSignal == "update" {
		if stateCheckSignal.Agents.DatadogAgent.Enabled {
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
					Version:       datadogAgent.Version,
					HostTags:      stateCheckSignal.HostTags,
					Hostname:      l.extractHostnameFromFiles(files),
				}
				action = act
			}
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

// start execute loop for adapter
func (l *ManagerAdapter) start() {
	l.logger.Debug("start loop", "trace", "agent-os-instance.manager_adapter.start")
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
