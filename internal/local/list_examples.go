package local

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/cloudfoundry/cloud-service-broker/pkg/client"
)

func ListExamples(cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker := startBroker(pakDir)
	defer broker.Stop()

	examples, err := listExamples(broker.Port)
	if err != nil {
		log.Fatal(err)
	}

	tp := newTablePrinter("Example Name", "Service Offering Name")
	for _, e := range examples {
		tp.row(e.Name, e.ServiceName)
	}
	tp.print()
}

func listExamples(port int) ([]client.CompleteServiceExample, error) {
	response, err := http.Get(fmt.Sprintf("http://localhost:%d/examples", port))
	if err != nil {
		return nil, fmt.Errorf("error getting examples from broker: %w", err)
	}

	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading body: %w", err)
	}

	var examples []client.CompleteServiceExample
	if err := json.Unmarshal(data, &examples); err != nil {
		return nil, fmt.Errorf("error unmarshalling JSON response: %w", err)
	}

	return examples, nil
}
