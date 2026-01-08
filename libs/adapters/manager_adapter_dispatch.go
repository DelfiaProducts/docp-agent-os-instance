package adapters

import (
	"context"
	"path/filepath"
	"time"

	"github.com/OryaHub/agent-os-instance/libs/dto"
	"github.com/OryaHub/agent-os-instance/libs/utils"
)

// ExecuteAuthCall execute call to auth and save access token received
func (l *ManagerAdapter) ExecuteAuthCall() error {
	l.logger.Info("execute auth call", "timestamp", time.Now())
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
		if err := l.json.Unmarshall(resp, &authResponse); err != nil {
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
	//get tracer id from state check response
	signalReceivedBytes, err := l.GetStateReceived()
	if err != nil {
		return err
	}
	var signalResponse dto.StateCheckResponse
	if err := l.json.Unmarshall(signalReceivedBytes, &signalResponse); err != nil {
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
			//populate tracer id if exist
			if signalResponse.TraceId != "" {
				transactionStatus.TraceId = signalResponse.TraceId
			}
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

// Close closing collect loop
func (l *ManagerAdapter) Close() error {
	l.logger.Info("execute close", "timestamp", time.Now().UTC())
	l.logger.Debug("execute close", "trace", "agent-os-instance.manager_adapter.Close")
	l.chanClose <- struct{}{}
	l.wg.Add(1)
	go l.closeChannels()
	return nil
}

// HandlerSCMManager execute handler for scm manager
func (l *ManagerAdapter) HandlerSCMManager() error {
	if err := l.osOperation.Execute(); err != nil {
		return err
	}
	return nil
}
