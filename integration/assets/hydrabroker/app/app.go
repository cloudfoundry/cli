package app

import (
	"net/http"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"
	"github.com/gorilla/mux"
)

func App() *mux.Router {
	s := store.New()

	handle := func(handler func(s *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := handler(s, w, r); err != nil {
				handleError(w, r, err)
			}
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")

	r.HandleFunc("/config", handle(configCreateBroker)).Methods("POST")
	r.HandleFunc("/config", handle(configListBrokers)).Methods("GET")
	r.HandleFunc("/config/{guid}", handle(configDeleteBroker)).Methods("DELETE")
	r.HandleFunc("/config/{guid}", handle(configRecreateBroker)).Methods("PUT")

	r.HandleFunc("/broker/{guid}/v2/catalog", handle(brokerCatalog)).Methods("GET")
	r.HandleFunc("/broker/{guid}/v2/service_instances/{si_guid}", handle(brokerProvision)).Methods("PUT")
	r.HandleFunc("/broker/{guid}/v2/service_instances/{si_guid}", handle(brokerDeprovision)).Methods("DELETE")
	r.HandleFunc("/broker/{guid}/v2/service_instances/{si_guid}/last_operation", handle(brokerLastOperation)).Methods("GET")

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	switch e := err.(type) {
	case notFoundError:
		http.NotFound(w, r)
	case interface{ StatusCode() int }:
		http.Error(w, err.Error(), e.StatusCode())
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
