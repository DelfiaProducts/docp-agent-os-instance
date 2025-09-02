package operators

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/DelfiaProducts/docp-agent-os-instance/api"
	adapters "github.com/DelfiaProducts/docp-agent-os-instance/libs/adapters"
	libinterfaces "github.com/DelfiaProducts/docp-agent-os-instance/libs/interfaces"
	libutils "github.com/DelfiaProducts/docp-agent-os-instance/libs/utils"
)

// AgentOperator is struct for operator the linux
type AgentOperator struct {
	chanMetadata chan []byte
	chanErrors   chan error
	adapter      *adapters.AgentAdapter
	logger       libinterfaces.ILogger
	api          *api.DocpApi
	wg           *sync.WaitGroup
}

// NewAgentOperator return instance of linux operator
func NewAgentOperator() *AgentOperator {
	return &AgentOperator{
		chanMetadata: make(chan []byte, 1),
		chanErrors:   make(chan error, 1),
		wg:           &sync.WaitGroup{},
	}
}

// Setup configure operator
func (l *AgentOperator) Setup() error {
	port, err := libutils.GetPortAgentApi()
	if err != nil {
		return err
	}
	var logger libinterfaces.ILogger
	if runtime.GOOS == "windows" {
		workdir, err := libutils.GetWorkDirPath()
		if err != nil {
			return err
		}
		logPath := filepath.Join(workdir, "logs", "agent.log")
		loggerFile := libutils.NewDocpLoggerWindowsFileText(logPath)
		logger = loggerFile
	} else {
		logger = libutils.NewDocpLoggerJSON(os.Stdout)
	}
	l.logger = logger
	api := api.NewDocpApi(port, l.logger)
	if err := api.Setup(); err != nil {
		return err
	}
	l.api = api
	adapterAgent := adapters.NewAgentAdapter(l.logger)
	if err := adapterAgent.Prepare(); err != nil {
		return err
	}
	l.adapter = adapterAgent
	return nil
}

// consumerErrors execute consumer for errors
func (l *AgentOperator) consumerErrors() {
	l.logger.Debug("execute consumer errors", "trace", "docp-agent-os-instance.agent_operator.consumerErrors")
	defer l.wg.Done()
	for {
		select {
		case err, ok := <-l.chanErrors:
			l.logger.Debug("consumer errors", "ok", ok)
			if !ok {
				return
			}
			if err != nil {
				l.logger.Error("error received in consumer errors", "trace", "docp-agent-os-instance.agent_operator.consumerErrors", "error", err.Error())
			}
		}
	}
}

// apiListen is execute listen the api
func (l *AgentOperator) apiListen() {
	l.logger.Debug("api listen", "trace", "docp-agent-os-instance.agent_operator.apiListen")
	defer l.wg.Done()
	if err := l.api.Run(); err != nil {
		l.chanErrors <- err
	}
}

// Run execut loop the operator
func (l *AgentOperator) Run() error {
	if err := l.Setup(); err != nil {
		return err
	}
	l.logger.Debug("execute run", "trace", "docp-agent-os-instance.agent_operator.Run")
	l.logger.Info("execute agent")
	defer l.logger.Close()
	l.wg.Add(3)
	go l.comunicateSCM()
	go l.consumerErrors()
	go l.apiListen()
	l.wg.Wait()
	return nil
}

func (l *AgentOperator) comunicateSCM() {
	defer l.wg.Done()
	if err := l.adapter.HandlerSCMManager(); err != nil {
		l.chanErrors <- err
	}
}
