//go:build windows

package components

import (
	"time"

	"golang.org/x/sys/windows/svc"
)

type HandlerSCM struct{}

func NewHandlerSCM() *HandlerSCM {
	return &HandlerSCM{}
}

func (h *HandlerSCM) Execute(args []string, changes <-chan svc.ChangeRequest, status chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	status <- svc.Status{State: svc.StartPending}
	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for {
		select {
		case change := <-changes:
			if change.Cmd == svc.Stop || change.Cmd == svc.Shutdown {
				status <- svc.Status{State: svc.StopPending}
				return false, 0
			}
		default:
			time.Sleep(5 * time.Second)
		}
	}
}

type SCMManager struct {
	serviceName string
	handlerSCM  *HandlerSCM
}

func NewSCMManager(serviceName string, handlerSCM *HandlerSCM) *SCMManager {
	return &SCMManager{
		serviceName: serviceName,
		handlerSCM:  handlerSCM,
	}
}

func (s *SCMManager) Run() error {
	isService, err := svc.IsWindowsService()
	if err != nil {
		return err
	}

	if isService {
		if err := svc.Run(s.serviceName, s.handlerSCM); err != nil {
			return err
		}
	}

	return nil
}
