//go:build windows
// +build windows

package main

import (
	"log"
	"slate-nexus-agent/logger"

	"golang.org/x/sys/windows/svc"
)

type Service struct{}

func (s *Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	stop := make(chan struct{})

	go runAgent(stop)

	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				close(stop)
				changes <- svc.Status{State: svc.StopPending}
				return
			default:
				log.Printf("unexpected control request: #%d", c)
			}
		}
	}
}

func runAsService() {
	err := svc.Run("SlateNexusAgent", &Service{})
	if err != nil {
		logger.LogError("Service failed: %v", err)
	}
}
