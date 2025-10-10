package operators

// consumerErrors execute consumer for errors
func (l *AgentOperator) consumerErrors() {
	l.logger.Debug("execute consumer errors", "trace", "agent-os-instance.agent_operator.consumerErrors")
	defer l.wg.Done()
	for {
		select {
		case err, ok := <-l.chanErrors:
			l.logger.Debug("consumer errors", "ok", ok)
			if !ok {
				return
			}
			if err != nil {
				l.logger.Error("error received in consumer errors", "trace", "agent-os-instance.agent_operator.consumerErrors", "error", err.Error())
			}
		}
	}
}

// apiListen is execute listen the api
func (l *AgentOperator) apiListen() {
	l.logger.Debug("api listen", "trace", "agent-os-instance.agent_operator.apiListen")
	defer l.wg.Done()
	if err := l.api.Run(); err != nil {
		l.chanErrors <- err
	}
}

func (l *AgentOperator) comunicateSCM() {
	defer l.wg.Done()
	if err := l.adapter.HandlerSCMManager(); err != nil {
		l.chanErrors <- err
	}
}
