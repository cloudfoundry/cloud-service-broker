package helper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/onsi/gomega"
)

func waitForBrokerToStart(port int) {
	ping := func() (*http.Response, error) {
		return http.Head(fmt.Sprintf("http://localhost:%d", port))
	}

	gomega.Eventually(ping, 30*time.Second).Should(gomega.HaveHTTPStatus(http.StatusOK))
}
