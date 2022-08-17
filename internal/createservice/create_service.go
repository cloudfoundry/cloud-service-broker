// Package createservice is an experimental mimic for the "cf create-service" command
package createservice

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/pkg/brokerpak"
	"github.com/cloudfoundry/cloud-service-broker/pkg/client"
	"github.com/cloudfoundry/cloud-service-broker/utils/freeport"
	"github.com/pborman/uuid"
	"github.com/pivotal-cf/brokerapi/v8/domain"
)

func Run(service, plan, name, c, cachePath string) {
	pakPath := pack(cachePath)
	port := freeport.Must()
	username := uuid.New()
	password := uuid.New()
	workdir, err := os.MkdirTemp("", "csb-*")
	if err != nil {
		panic(err)
	}
	pakPath, err = filepath.Abs(pakPath)
	if err != nil {
		panic(err)
	}
	os.Symlink(pakPath, filepath.Join(workdir, filepath.Base(pakPath)))
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	dbfile := filepath.Join(cwd, ".csb.db")
	h := start(username, password, workdir, dbfile, port)
	defer h.Terminate()

	wait(port, h)

	clnt, err := client.New(username, password, "localhost", port)
	if err != nil {
		panic(err)
	}
	serviceID, planID := lookupServiceInfo(clnt, service, plan)
	instanceID := NameToID(name)
	provision(clnt, instanceID, serviceID, planID)
	fmt.Printf("**** HERE 2 %s %s\n", serviceID, planID)

	for start := time.Now(); lastOperation(clnt, instanceID); {
		if time.Since(start) > time.Hour {
			panic("timed out")
		}
		time.Sleep(time.Second)
	}
}

func lastOperation(clnt *client.Client, instanceID string) bool {
	lastOperationResponse := clnt.LastOperation(instanceID, uuid.New())
	switch {
	case lastOperationResponse.Error != nil:
		panic(lastOperationResponse.Error)
	case lastOperationResponse.StatusCode != http.StatusOK:
		panic(fmt.Sprintf("unexpected last operation response: %d", lastOperationResponse.StatusCode))
	}

	var lo domain.LastOperation
	if err := json.Unmarshal(lastOperationResponse.ResponseBody, &lo); err != nil {
		panic(err)
	}

	switch lo.State {
	case domain.InProgress:
		return true
	case domain.Succeeded:
		return false
	default:
		panic(fmt.Sprintf("provision failed: %s", lo.Description))
	}
}

func provision(clnt *client.Client, instanceID, serviceID, planID string) {
	provisionResponse := clnt.Provision(instanceID, serviceID, planID, uuid.New(), nil)
	switch {
	case provisionResponse.Error != nil:
		panic(fmt.Sprintf("provision error: %s", provisionResponse.Error))
	case provisionResponse.StatusCode != http.StatusAccepted:
		panic(fmt.Sprintf("unexpected provision reponse: %d", provisionResponse.StatusCode))
	}
}

func lookupServiceInfo(clnt *client.Client, serviceName, planName string) (string, string) {
	catalogResponse := clnt.Catalog(uuid.New())
	switch {
	case catalogResponse.Error != nil:
		panic(catalogResponse.Error)
	case catalogResponse.StatusCode != http.StatusOK:
		panic(fmt.Sprintf("bad catalog response %d: %s", catalogResponse.StatusCode, catalogResponse.String()))
	}

	var catalog struct {
		Services []domain.Service `json:"services"`
	}
	if err := json.Unmarshal(catalogResponse.ResponseBody, &catalog); err != nil {
		panic(err)
	}

	for _, s := range catalog.Services {
		if s.Name == serviceName {
			for _, p := range s.Plans {
				if p.Name == planName {
					return s.ID, p.ID
				}
			}
			panic(fmt.Sprintf("could not find plan %q in service %q", planName, serviceName))
		}
	}
	panic(fmt.Sprintf("could not find service %q in catalog", serviceName))
}

func newHandle(cmd *exec.Cmd) *handle {
	h := handle{cmd: cmd}
	go func() {
		if err := cmd.Wait(); err != nil {
			panic(fmt.Sprintf("server process failed: %s", err))
		}
		h.Exited = true
	}()
	return &h
}

type handle struct {
	cmd    *exec.Cmd
	Exited bool
}

func (h *handle) Terminate() {
	h.cmd.Process.Kill()
}

func start(username, password, workdir, dbfile string, port int) *handle {
	cmd := exec.Command(os.Args[0], "serve")
	cmd.Dir = workdir
	cmd.Env = append(
		os.Environ(),
		"CSB_LISTENER_HOST=localhost",
		"DB_TYPE=sqlite3",
		fmt.Sprintf("DB_PATH=%s", dbfile),
		fmt.Sprintf("PORT=%d", port),
		fmt.Sprintf("SECURITY_USER_NAME=%s", username),
		fmt.Sprintf("SECURITY_USER_PASSWORD=%s", password),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	return newHandle(cmd)
}

func wait(port int, h *handle) {
	for start := time.Now(); time.Since(start) < 5*time.Minute && !h.Exited; {
		res, err := http.Get(fmt.Sprintf("http://localhost:%d//info", port))
		if err == nil && res.StatusCode == http.StatusOK {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	panic("timed out or died")
}

func pack(cachePath string) string {
	pakPath, err := brokerpak.Pack("", cachePath, false)
	if err != nil {
		log.Fatalf("error while packing: %v", err)
	}

	if err := brokerpak.Validate(pakPath); err != nil {
		log.Fatalf("created: %v, but it failed validity checking: %v\n", pakPath, err)
	} else {
		fmt.Printf("created: %v\n", pakPath)
	}

	return pakPath
}
