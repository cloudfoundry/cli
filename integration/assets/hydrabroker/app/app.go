package app

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"
	"github.com/gorilla/mux"
)

func App() *mux.Router {
	s := store.New()

	handle := func(name string, handler func(s *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
		return func(w http.ResponseWriter, r *http.Request) {
			if err := handler(s, w, r); err != nil {
				http.Error(w, fmt.Sprintf("Failed in handler %s: %s", name, err.Error()), http.StatusBadRequest)
			}
		}
	}

	brokerAuthMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := checkAuth(s, w, r); err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	r := mux.NewRouter()
	r.HandleFunc("/", aliveness).Methods("HEAD", "GET")

	r.HandleFunc("/config", handle("create", configCreateBroker)).Methods("POST")
	r.HandleFunc("/config", handle("list", configListBrokers)).Methods("GET")
	r.HandleFunc("/config/{guid}", handle("delete", configDeleteBroker)).Methods("DELETE")

	b := r.PathPrefix("/broker/{guid}").Subrouter()
	b.Use(brokerAuthMiddleware)
	b.HandleFunc("/v2/catalog", handle("catalog", brokerCatalog)).Methods("GET")
	b.HandleFunc("/v2/service_instances/{si_guid}", handle("provision", brokerProvision)).Methods("PUT")

	return r
}

func aliveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
