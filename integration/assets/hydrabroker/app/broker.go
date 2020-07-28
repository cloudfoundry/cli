package app

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/resources"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
)

type requestGUIDs struct {
	brokerGUID          string
	serviceInstanceGUID string
	bindingGUID         string
}

func brokerParseHeaders(store *store.BrokerConfigurationStore, r *http.Request) (config.BrokerConfiguration, requestGUIDs, error) {
	guids, err := readGUIDs(r)
	if err != nil {
		return config.BrokerConfiguration{}, requestGUIDs{}, err
	}

	cfg, ok := store.GetBrokerConfiguration(guids.brokerGUID)
	if !ok {
		return config.BrokerConfiguration{}, requestGUIDs{}, notFoundError{}
	}

	givenUsername, givenPassword, ok := r.BasicAuth()
	if !ok {
		return config.BrokerConfiguration{}, requestGUIDs{}, unauthorizedError{}
	}

	// Compare everything every time to protect against timing attacks
	if 2 != subtle.ConstantTimeCompare([]byte(cfg.Username), []byte(givenUsername))+
		subtle.ConstantTimeCompare([]byte(cfg.Password), []byte(givenPassword)) {
		return config.BrokerConfiguration{}, requestGUIDs{}, unauthorizedError{}
	}

	return cfg, guids, nil
}

func brokerCatalog(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	config, _, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	if config.CatalogResponse != 0 {
		w.WriteHeader(config.CatalogResponse)
		return nil
	}

	var services []domain.Service
	for _, ser := range config.Services {
		var plans []domain.ServicePlan
		s := ser // Copy to protect from memory reuse

		for _, pla := range s.Plans {
			var mi *domain.MaintenanceInfo
			p := pla // Copy to protect from memory reuse

			if p.MaintenanceInfo != nil {
				mi = &domain.MaintenanceInfo{
					Version:     p.MaintenanceInfo.Version,
					Description: p.MaintenanceInfo.Description,
				}
			}

			var costs []domain.ServicePlanCost
			if p.Costs != nil {
				for _, c := range p.Costs {
					costs = append(costs, domain.ServicePlanCost{
						Amount: c.Amount,
						Unit:   c.Unit,
					})
				}
			}

			plans = append(plans, domain.ServicePlan{
				Name:            p.Name,
				ID:              p.ID,
				Description:     p.Description,
				MaintenanceInfo: mi,
				Bindable:        &s.Bindable,
				Free:            &p.Free,
				Metadata: &domain.ServicePlanMetadata{
					Costs: costs,
				},
			})
		}
		services = append(services, domain.Service{
			Name:                 s.Name,
			ID:                   s.ID,
			Description:          s.Description,
			Tags:                 s.Tags,
			Plans:                plans,
			BindingsRetrievable:  s.Bindable,
			InstancesRetrievable: s.InstancesRetrievable,
			Requires:             brokerCastRequires(s.Requires),
			Metadata: &domain.ServiceMetadata{
				Shareable:        &s.Shareable,
				DocumentationUrl: s.DocumentationURL,
			},
		})
	}

	return respondWithJSON(w, apiresponses.CatalogResponse{Services: services})
}

func brokerProvision(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	if config.ProvisionResponse != 0 {
		w.WriteHeader(config.ProvisionResponse)
		return nil
	}

	var details resources.ServiceInstanceDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		return newBadRequestError("invalid JSON", err)
	}

	if err := store.CreateServiceInstance(guids.brokerGUID, guids.serviceInstanceGUID, details); err != nil {
		return err
	}

	response := map[string]interface{}{
		"dashboard_url": `http://example.com`,
	}

	switch config.AsyncResponseDelay {
	case 0:
		w.WriteHeader(http.StatusCreated)
		return respondWithJSON(w, response)
	default:
		return brokerAsyncResponse(w, r, config.AsyncResponseDelay, response)
	}
}

func brokerRetrieve(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	_, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	details, err := store.RetrieveServiceInstance(guids.brokerGUID, guids.serviceInstanceGUID)
	if err != nil {
		return err
	}

	parameters := details.Parameters
	if parameters == nil { // Ensure response contains `{}` rather than `null`
		parameters = make(map[string]interface{})
	}

	response := map[string]interface{}{
		"parameters": parameters,
	}

	w.WriteHeader(http.StatusOK)
	return respondWithJSON(w, response)
}

func brokerUpdate(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	if config.UpdateResponse != 0 {
		w.WriteHeader(config.UpdateResponse)
		return nil
	}

	_, err = store.RetrieveServiceInstance(guids.brokerGUID, guids.serviceInstanceGUID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	var details resources.ServiceInstanceDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		return newBadRequestError("invalid JSON", err)
	}

	if err := store.CreateServiceInstance(guids.brokerGUID, guids.serviceInstanceGUID, details); err != nil {
		return err
	}

	switch config.AsyncResponseDelay {
	case 0:
		w.WriteHeader(http.StatusOK)
		return respondWithJSON(w, map[string]interface{}{})
	default:
		return brokerAsyncResponse(w, r, config.AsyncResponseDelay, nil)
	}
}

func brokerDeprovision(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	config, _, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	if config.DeprovisionResponse != 0 {
		w.WriteHeader(config.DeprovisionResponse)
		return nil
	}

	switch config.AsyncResponseDelay {
	case 0:
		w.WriteHeader(http.StatusOK)
		return respondWithJSON(w, map[string]interface{}{})
	default:
		return brokerAsyncResponse(w, r, config.AsyncResponseDelay, nil)
	}
}

func brokerBind(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	if config.BindResponse != 0 {
		w.WriteHeader(config.BindResponse)
		return nil
	}

	_, err = store.RetrieveServiceInstance(guids.brokerGUID, guids.serviceInstanceGUID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	switch config.AsyncResponseDelay {
	case 0:
		w.WriteHeader(http.StatusCreated)
		return respondWithJSON(w, map[string]interface{}{})
	default:
		return brokerAsyncResponse(w, r, config.AsyncResponseDelay, nil)
	}
}

func brokerGetBinding(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	config, _, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	if config.GetBindingResponse != 0 {
		w.WriteHeader(config.GetBindingResponse)
		return nil
	}

	return respondWithJSON(w, map[string]interface{}{})
}

func brokerUnbind(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	config, _, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	if config.UnbindResponse != 0 {
		w.WriteHeader(config.UnbindResponse)
		return nil
	}

	switch config.AsyncResponseDelay {
	case 0:
		w.WriteHeader(http.StatusOK)
		return respondWithJSON(w, map[string]interface{}{})
	default:
		return brokerAsyncResponse(w, r, config.AsyncResponseDelay, nil)
	}
}

func brokerAsyncResponse(w http.ResponseWriter, r *http.Request, duration time.Duration, params map[string]interface{}) error {
	if r.FormValue("accepts_incomplete") != "true" {
		return fmt.Errorf("want to respond async, but got `accepts_incomplete` = `%v`", r.FormValue("accepts_incomplete"))
	}

	if params == nil {
		params = make(map[string]interface{})
	}

	params["operation"] = time.Now().Add(duration)

	w.WriteHeader(http.StatusAccepted)
	return respondWithJSON(w, params)
}

func brokerLastOperation(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	_, _, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	var when time.Time
	if err := when.UnmarshalJSON([]byte(`"` + r.FormValue("operation") + `"`)); err != nil {
		return err
	}

	result := apiresponses.LastOperationResponse{
		Description: "very happy service",
	}
	switch time.Now().After(when) {
	case true:
		result.State = domain.Succeeded
	case false:
		result.State = domain.InProgress
	}

	return respondWithJSON(w, result)
}

func brokerCastRequires(input []string) (result []domain.RequiredPermission) {
	for _, v := range input {
		result = append(result, domain.RequiredPermission(v))
	}

	return
}
