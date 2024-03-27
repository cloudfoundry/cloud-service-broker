package tfstatebackend

import (
	"fmt"
	"io"
	"net/http"

	"code.cloudfoundry.org/lager/v3"
	"github.com/cloudfoundry/cloud-service-broker/v2/internal/storage"
	"github.com/cloudfoundry/cloud-service-broker/v2/pkg/providers/tf/workspace"
)

type backend struct {
	store  *storage.Storage
	logger lager.Logger
}

var (
	StateBackendPort string
)

func New(store *storage.Storage, logger lager.Logger) *http.ServeMux {
	b := backend{store, logger}

	router := http.NewServeMux()
	router.HandleFunc("/{id}", func(w http.ResponseWriter, r *http.Request) {
		b.logger.Info("handle-tf-state-store-request", lager.Data{"method": r.Method, "uri": r.RequestURI})
		switch r.Method {
		// case "LOCK":
		// 	b.HandleLockState(w, r)
		// case "UNLOCK":
		// 	b.HandleUnlockState(w, r)
		case http.MethodGet:
			b.HandleGetState(w, r)
		case http.MethodPost:
			b.HandleUpdateState(w, r)
		case http.MethodDelete:
			b.HandleDeleteState(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return router
}

func (b *backend) HandleGetState(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		b.logger.Error("get-terraform-deployment", fmt.Errorf("could not extract id from url"), lager.Data{"url": r.RequestURI})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	deployment, err := b.store.GetTerraformDeployment(id)
	if err != nil {
		b.logger.Error("get-terraform-deployment", err, lager.Data{"id": id})
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if len(deployment.TFWorkspace().State) == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(deployment.TFWorkspace().State)
}

func (b *backend) HandleUpdateState(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		b.logger.Error("get-terraform-deployment", fmt.Errorf("could not extract id from url"), lager.Data{"url": r.RequestURI})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	deployment, err := b.store.GetTerraformDeployment(id)
	if err != nil {
		b.logger.Error(fmt.Sprintf("failed to get terraform deployment: %s", id), err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	state, err := io.ReadAll(r.Body)
	if err != nil {
		b.logger.Error(fmt.Sprintf("failed to read body update request with id: %s", id), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err := workspace.NewTfstate(state); err != nil {
		b.logger.Error(fmt.Sprintf("failed invalid state for id: %s", id), err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	deployment.TFWorkspace().State = state
	if err := b.store.StoreTerraformDeployment(deployment); err != nil {
		b.logger.Error(fmt.Sprintf("failed to store state for id: %s", id), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Println("Update State (deployment):")
	fmt.Println(string(deployment.TFWorkspace().State))
	w.WriteHeader(http.StatusOK)
}

func (b *backend) HandleDeleteState(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id != "" {
		b.logger.Error("failed to get terraform deployment", fmt.Errorf("could not extract deployment id from: %s", r.RequestURI))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	deployment, err := b.store.GetTerraformDeployment(id)
	if err != nil {
		b.logger.Error(fmt.Sprintf("failed to get terraform deployment: %s", id), err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	deployment.TFWorkspace().State = make([]byte, 0)
	if err := b.store.StoreTerraformDeployment(deployment); err != nil {
		b.logger.Error(fmt.Sprintf("failed to store state for id: %s", id), err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
