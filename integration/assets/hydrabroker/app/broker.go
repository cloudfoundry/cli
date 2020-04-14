package app

import (
	"crypto/subtle"
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
		return notFoundError{}
	}

	givenUsername, givenPassword, ok := r.BasicAuth()
	if !ok {
		return unauthorizedError{}
	}

	// Compare everything every time to protect against timing attacks
	if 2 == subtle.ConstantTimeCompare([]byte(config.Username), []byte(givenUsername))+
		subtle.ConstantTimeCompare([]byte(config.Password), []byte(givenPassword)) {
		return nil
	}

	return unauthorizedError{}
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

	if config.CatalogResponse != 0 {
		w.WriteHeader(config.CatalogResponse)
		return nil
	}

	var services []domain.Service
	for _, s := range config.Services {
		var plans []domain.ServicePlan
		for _, p := range s.Plans {
			var mi *domain.MaintenanceInfo

			if p.MaintenanceInfo != nil {
				mi = &domain.MaintenanceInfo{
					Version:     p.MaintenanceInfo.Version,
					Description: p.MaintenanceInfo.Description,
				}
			}

			plans = append(plans, domain.ServicePlan{
				Name:            p.Name,
				ID:              p.ID,
				Description:     p.Description,
				MaintenanceInfo: mi,
			})
		}
		services = append(services, domain.Service{
			Name:        s.Name,
			ID:          s.ID,
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

	config, ok := store.GetBrokerConfiguration(guid)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	if config.ProvisionResponse != 0 {
		w.WriteHeader(config.ProvisionResponse)
		return nil
	}

	w.WriteHeader(http.StatusCreated)
	return respondWithJSON(w, map[string]interface{}{})
}

func brokerDeprovision(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	guid, err := readGUID(r)
	if err != nil {
		return err
	}

	config, ok := store.GetBrokerConfiguration(guid)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	if config.DeprovisionResponse != 0 {
		w.WriteHeader(config.DeprovisionResponse)
		return nil
	}

	w.WriteHeader(http.StatusOK)
	return respondWithJSON(w, map[string]interface{}{})
}
