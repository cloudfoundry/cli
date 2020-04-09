package app

import (
	"errors"
	"net/http"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
)

func checkAuth(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	guid, err := readGUID(r)
	if err != nil {
		return err
	}

	config, ok := store.GetBrokerConfiguration(guid)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	givenUsername, givenPassword, ok := r.BasicAuth()
	if !ok || config.Username != givenUsername || config.Password != givenPassword {
		return errors.New("not authorized")
	}

	return nil
}

func brokerCatalog(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	guid, err := readGUID(r)
	if err != nil {
		return err
	}

	config, ok := store.GetBrokerConfiguration(guid)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	var services []domain.Service
	for _, s := range config.Services {
		var plans []domain.ServicePlan
		for _, p := range s.Plans {
			plans = append(plans, domain.ServicePlan{
				Name:        p.Name,
				ID:          p.GUID,
				Description: p.Description,
			})
		}
		services = append(services, domain.Service{
			Name:        s.Name,
			ID:          s.GUID,
			Description: s.Description,
			Plans:       plans,
		})
	}

	return respondWithJSON(w, apiresponses.CatalogResponse{Services: services})
}

func brokerProvision(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	guid, err := readGUID(r)
	if err != nil {
		return err
	}

	_, ok := store.GetBrokerConfiguration(guid)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	w.WriteHeader(http.StatusCreated)
	return respondWithJSON(w, map[string]interface{}{})
}
