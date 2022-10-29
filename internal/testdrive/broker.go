package testdrive

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/utils/freeport"
	"github.com/pborman/uuid"
)

type Broker struct {
	Database Database
	Port     int
	username string
	password string
	cancel   context.CancelFunc
}

func StartBroker(csbPath string, bpk Brokerpak, db Database) (*Broker, error) {
	port, err := freeport.Port()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	broker := Broker{
		Database: db,
		Port:     port,
		username: uuid.New(),
		password: uuid.New(),
		cancel:   cancel,
	}

	cmd := exec.CommandContext(ctx, csbPath)
	cmd.Dir = string(bpk)
	cmd.Env = append(
		os.Environ(),
		"CSB_LISTENER_HOST=localhost",
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", broker.Database),
		fmt.Sprintf("PORT=%d", broker.Port),
		fmt.Sprintf("SECURITY_USER_NAME=%s", broker.username),
		fmt.Sprintf("SECURITY_USER_PASSWORD=%s", broker.password),
	)

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	const timeout = time.Minute
	for start := time.Now(); time.Since(start) < timeout; {
		response, err := http.Head(fmt.Sprintf("http://localhost:%d", port))
		if err == nil && response.StatusCode == http.StatusOK {
			return &broker, nil
		}
	}

	cancel()
	return nil, fmt.Errorf("timed out after %s waiting for broker to start", timeout)
}

func (b *Broker) Stop() {
	b.cancel()
}
