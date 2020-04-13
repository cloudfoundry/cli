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

	brokerAuthMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := checkAuth(s, w, r); err != nil {
				handleError(w, r, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")

	r.HandleFunc("/config", handle(configCreateBroker)).Methods("POST")
	r.HandleFunc("/config", handle(configListBrokers)).Methods("GET")
	r.HandleFunc("/config/{guid}", handle(configDeleteBroker)).Methods("DELETE")
	r.HandleFunc("/config/{guid}", handle(configRecreateBroker)).Methods("PUT")

	b := r.PathPrefix("/broker/{guid}").Subrouter()
	b.Use(brokerAuthMiddleware)
	b.HandleFunc("/v2/catalog", handle(brokerCatalog)).Methods("GET")
	b.HandleFunc("/v2/service_instances/{si_guid}", handle(brokerProvision)).Methods("PUT")
	b.HandleFunc("/v2/service_instances/{si_guid}", handle(brokerDeprovision)).Methods("DELETE")

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
