package testdrive

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	//lint:ignore ST1001 we do not care because this is a test helper
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/client"
	"github.com/cloudfoundry/cloud-service-broker/v2/utils/freeport"
	"github.com/google/uuid"
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

	username := uuid.NewString()
	password := uuid.NewString()

	cmd := exec.Command(csbPath, "serve")
	cmd.Dir = bpk
	cmd.Env = append(
		cfg.env,
		"CSB_LISTENER_HOST=localhost",
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", db),
		fmt.Sprintf("PORT=%d", port),
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
		Username: username,
		Password: password,
		runner:   newCommand(cmd),
		Stdout:   &stdout,
		Stderr:   &stderr,
	}

	start := time.Now()

	scheme := "http"

	for _, envVar := range cmd.Env {
		if strings.HasPrefix(envVar, "TLS_") {

			ignoreSelfSignedCerts := &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			}
			broker.Client.BaseURL.Scheme = "https"
			broker.Client.Transport = ignoreSelfSignedCerts
			scheme = "https"
			break
		}
	}

	for {
		response, err := broker.Client.Head(fmt.Sprintf("%s://localhost:%d", scheme, port))
		switch {
		case err == nil && response.StatusCode == http.StatusOK:
			return &broker, nil
		case time.Since(start) > time.Minute:
			if err := broker.runner.forceStop(); err != nil {
				return nil, err
			}
			return nil, fmt.Errorf("timed out after %s waiting for broker to start: %s\n%s", time.Since(start), stdout.String(), stderr.String())
		case broker.runner.exited:
			return nil, fmt.Errorf("failed to start broker: %w\n%s\n%s", broker.runner.err, stdout.String(), stderr.String())
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func createCAKeyPair(msg string) (*x509.Certificate, *rsa.PrivateKey) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Country: []string{msg},
		},
		IsCA:                  true,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).NotTo(HaveOccurred())

	return ca, caPrivKey
}

func createKeyPairSignedByCA(ca *x509.Certificate, caPrivKey *rsa.PrivateKey) ([]byte, []byte) {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Country: []string{"GB"},
		},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(10, 0, 0),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	Expect(err).NotTo(HaveOccurred())

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	Expect(err).NotTo(HaveOccurred())

	return encodeKeyPair(certBytes, x509.MarshalPKCS1PrivateKey(certPrivKey))
}

func encodeKeyPair(caBytes, caPrivKeyBytes []byte) ([]byte, []byte) {
	caPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: caPrivKeyBytes,
	})

	return caPEM, caPrivKeyPEM
}

func WithInvalidTLSConfig() StartBrokerOption {
	return func(cfg *startBrokerConfig) {
		tlsConfig(cfg, false)
	}
}

func WithTLSConfig() StartBrokerOption {
	return func(cfg *startBrokerConfig) {
		tlsConfig(cfg, true)
	}
}

func tlsConfig(cfg *startBrokerConfig, valid bool) {
	ca, caPrivKey := createCAKeyPair("US")

	serverCert, serverPrivKey := createKeyPairSignedByCA(ca, caPrivKey)

	certFileBuf, err := os.CreateTemp("", "")
	Expect(err).NotTo(HaveOccurred())
	defer certFileBuf.Close()

	privKeyFileBuf, err := os.CreateTemp("", "")
	Expect(err).NotTo(HaveOccurred())
	defer privKeyFileBuf.Close()

	if !valid {
		// If the isValid parameter is false, the server private key is intentionally corrupted
		// by modifying one of its bytes.
		serverPrivKey[10] = 'a'
	}

	Expect(os.WriteFile(privKeyFileBuf.Name(), serverPrivKey, 0o644)).To(Succeed())

	Expect(os.WriteFile(certFileBuf.Name(), serverCert, 0o644)).To(Succeed())

	cfg.env = append(cfg.env, fmt.Sprintf("TLS_CERT=%s", certFileBuf.Name()))
	cfg.env = append(cfg.env, fmt.Sprintf("TLS_PRIVATE_KEY=%s", privKeyFileBuf.Name()))
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
