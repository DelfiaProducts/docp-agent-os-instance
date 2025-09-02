package pkg

import (
	"context"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
)

// SystemdClient is struct for systemd client
type SystemdClient struct{}

// NewSystemdClient return instance of systemd client
func NewSystemdClient() *SystemdClient {
	return &SystemdClient{}
}

// Status verify status of service
func (s *SystemdClient) Status(service string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	unitStatus, err := conn.GetUnitPropertyContext(ctx, service, "ActiveState")
	if err != nil {
		return "", err
	}
	return unitStatus.Value.String(), nil
}

// AlreadyInstalledService return if already service installed
func (s *SystemdClient) AlreadyInstalledService(service string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	units, err := conn.ListUnitsByNamesContext(ctx, []string{service})
	if err != nil {
		return false, err
	}

	if len(units) == 0 || (len(units) == 1 && units[0].LoadState == "not-found") {
		return false, nil
	}

	return true, nil
}
