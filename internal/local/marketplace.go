// Package local is an experimental mimic for the "cf create-service" command
package local

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/cloudfoundry/cloud-service-broker/internal/testdrive"
)

func Marketplace(cachePath string) {
	pakDir, cleanup := pack(cachePath)
	defer cleanup()

	broker, err := testdrive.StartBroker(os.Args[0], pakDir, databasePath(), testdrive.WithOutputs(os.Stdout, os.Stderr))
	if err != nil {
		log.Fatal(err)
	}
	defer broker.Stop()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.StripEscape)
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Service Offering\tPlans")
	_, _ = fmt.Fprintln(w, "----------------\t-----")
	for _, s := range catalog(broker.Client) {
		var plans []string
		for _, p := range s.Plans {
			plans = append(plans, p.Name)
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\n", s.Name, strings.Join(plans, ", "))
	}
	_, _ = fmt.Fprintln(w)
	_ = w.Flush()
}
