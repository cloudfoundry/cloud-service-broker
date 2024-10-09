// Package testdrive is used in testing and local development to take the broker for a test drive
package testdrive

import (
	"bytes"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/client"
)

type Broker struct {
	Database string
	Port     int
	Client   *client.Client
	runner   *runner
	Username string
	Password string
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
}

// Stop requests that the broker stops and waits for it to exit
func (b *Broker) Stop() error {
	switch {
	case b == nil, b.runner == nil:
		return nil
	default:
		return b.runner.gracefulStop()
	}
}

// RequestStop requests that the broker stop, but does not wait for exit
func (b *Broker) RequestStop() error {
	switch {
	case b == nil, b.runner == nil:
		return nil
	default:
		return b.runner.requestStop()
	}
}

// Terminate forces the broker to stop
func (b *Broker) Terminate() error {
	switch {
	case b == nil, b.runner == nil:
		return nil
	default:
		return b.runner.forceStop()
	}
}
