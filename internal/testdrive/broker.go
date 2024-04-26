// Package testdrive is used in testing and local development to take the broker for a test drive
package testdrive

import (
	"bytes"

	"github.com/cloudfoundry/cloud-service-broker/v3/pkg/client"
)

type Broker struct {
	Database string
	Port     int
	Client   *client.Client
	runner   *runner
	username string
	password string
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
}

func (b *Broker) Stop() error {
	switch {
	case b == nil, b.runner == nil:
		return nil
	default:
		return b.runner.stop()
	}
}
