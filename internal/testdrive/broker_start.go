package testdrive

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/client"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/freeport"
	"github.com/pborman/uuid"
)

type StartBrokerOption func(config *startBrokerConfig)

type startBrokerConfig struct {
	env    []string
	stdout io.Writer
	stderr io.Writer
}

func StartBroker(csbPath, bpk, db string, opts ...StartBrokerOption) (*Broker, error) {
	var stdout, stderr bytes.Buffer
	cfg := startBrokerConfig{env: os.Environ()}

	for _, o := range opts {
		o(&cfg)
	}

	port, err := freeport.Port()
	if err != nil {
		return nil, err
	}

	backendport, err := freeport.Port()
	if err != nil {
		return nil, err
	}

	username := uuid.New()
	password := uuid.New()

	cmd := exec.Command(csbPath, "serve")
	cmd.Dir = bpk
	cmd.Env = append(
		cfg.env,
		"CSB_LISTENER_HOST=localhost",
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", db),
		fmt.Sprintf("PORT=%d", port),
		fmt.Sprintf("STATE_BACKEND_PORT=%d", backendport),
		fmt.Sprintf("SECURITY_USER_NAME=%s", username),
		fmt.Sprintf("SECURITY_USER_PASSWORD=%s", password),
	)

	switch cfg.stdout {
	case nil:
		cmd.Stdout = &stdout
	default:
		cmd.Stdout = io.MultiWriter(&stdout, cfg.stdout)
	}

	switch cfg.stderr {
	case nil:
		cmd.Stderr = &stderr
	default:
		cmd.Stderr = io.MultiWriter(&stderr, cfg.stderr)
	}

	clnt, err := client.New(username, password, "localhost", port)
	if err != nil {
		return nil, err
	}

	broker := Broker{
		Database: db,
		Port:     port,
		Client:   clnt,
		username: username,
		password: password,
		runner:   newCommand(cmd),
		Stdout:   &stdout,
		Stderr:   &stderr,
	}

	start := time.Now()
	for {
		response, err := http.Head(fmt.Sprintf("http://localhost:%d", port))
		switch {
		case err == nil && response.StatusCode == http.StatusOK:
			return &broker, nil
		case time.Since(start) > time.Minute:
			if err := broker.runner.stop(); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("timed out after %s waiting for broker to start: %s\n%s", time.Since(start), stdout.String(), stderr.String())
		case broker.runner.exited:
			return nil, fmt.Errorf("failed to start broker: %w\n%s\n%s", broker.runner.err, stdout.String(), stderr.String())
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func WithEnv(extraEnv ...string) StartBrokerOption {
	return func(cfg *startBrokerConfig) {
		cfg.env = append(cfg.env, extraEnv...)
	}
}

func WithAllowedEnvs(allowed []string) StartBrokerOption {
	a := make(map[string]struct{})
	for _, allow := range allowed {
		a[allow] = struct{}{}
	}

	return func(cfg *startBrokerConfig) {
		var result []string
		for _, e := range cfg.env {
			name := varname(e)
			if _, ok := a[name]; ok || strings.HasPrefix(name, "GSB_") {
				result = append(result, e)
			}
		}
		cfg.env = result
	}
}

func WithOutputs(stdout, stderr io.Writer) StartBrokerOption {
	return func(cfg *startBrokerConfig) {
		cfg.stdout = stdout
		cfg.stderr = stderr
	}
}

func varname(e string) string {
	parts := strings.SplitN(e, "=", 2)
	return parts[0]
}
