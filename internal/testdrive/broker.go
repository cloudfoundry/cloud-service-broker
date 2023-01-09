// Package testdrive is used in testing and local development to take the broker for a test drive
package testdrive

import (
	"bytes"

	"github.com/cloudfoundry/cloud-service-broker/pkg/client"
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
	switch b.runner {
	case nil:
		return nil
	default:
		return b.runner.stop()
	}
}
