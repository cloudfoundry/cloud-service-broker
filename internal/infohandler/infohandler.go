// Package infohandler handles the /info endpoint
package infohandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudfoundry/cloud-service-broker/v3/utils"
)

type Config struct {
	BrokerVersion string
	Uptime        func() time.Duration
}

func New(cfg Config) http.HandlerFunc {
	type payload struct {
		Version string `json:"version"`
		Uptime  string `json:"uptime"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(payload{
			Version: cfg.BrokerVersion,
			Uptime:  cfg.Uptime().String(),
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("error marshalling info payload: %s", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "text/html")
		_, err = w.Write(data)
		if err != nil {
			http.Error(w, fmt.Sprintf("error writing response: %s", err), http.StatusInternalServerError)
			return
		}
	}
}

func NewDefault() http.HandlerFunc {
	startTime := time.Now()
	return New(Config{
		BrokerVersion: utils.Version,
		Uptime:        func() time.Duration { return time.Since(startTime) },
	})
}
