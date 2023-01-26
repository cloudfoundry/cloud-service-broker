package local

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/tabwriter"

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

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.StripEscape)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Example Name\tService Offering Name")
	_, _ = fmt.Fprintln(w, "------------\t---------------------")
	for _, e := range examples {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", e.Name, e.ServiceName)
	}
	_, _ = fmt.Fprintln(w)
	_ = w.Flush()
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
