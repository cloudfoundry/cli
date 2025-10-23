package app

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/resources"
	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/store"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/pivotal-cf/brokerapi/v7/domain/apiresponses"
)

type requestGUIDs struct {
	brokerGUID          store.BrokerID
	serviceInstanceGUID store.InstanceID
	bindingGUID         store.BindingID
}

func brokerParseHeaders(store *store.Store, r *http.Request) (config.BrokerConfiguration, requestGUIDs, error) {
	guids, err := readGUIDs(r)
	if err != nil {
		return config.BrokerConfiguration{}, requestGUIDs{}, err
	}

	cfg, ok := store.RetrieveBroker(guids.brokerGUID)
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

func brokerCatalog(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	config, _, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	if config.CatalogResponse != 0 {
		w.WriteHeader(config.CatalogResponse)
		return nil
	}
	log.Printf("presenting catalog")

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
			PlanUpdatable:        s.PlanUpdatable,
			Requires:             brokerCastRequires(s.Requires),
			Metadata: &domain.ServiceMetadata{
				Shareable:        &s.Shareable,
				DocumentationUrl: s.DocumentationURL,
			},
		})
	}

	return respondWithJSON(w, apiresponses.CatalogResponse{Services: services})
}

func brokerProvisionInstance(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}
	log.Printf("provisioning service instance %s for broker %s", guids.serviceInstanceGUID, guids.brokerGUID)

	if config.ProvisionResponse != 0 {
		w.WriteHeader(config.ProvisionResponse)
		return nil
	}

	var details resources.ServiceInstanceDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		return newBadRequestError("invalid JSON", err)
	}

	if err := store.CreateInstance(guids.brokerGUID, guids.serviceInstanceGUID, details); err != nil {
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

func brokerRetrieveInstance(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	_, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}
	log.Printf("retrieving service instance %s for broker %s", guids.serviceInstanceGUID, guids.brokerGUID)

	details, err := store.RetrieveInstance(guids.brokerGUID, guids.serviceInstanceGUID)
	if err != nil {
		return notFoundError{}
	}

	response := map[string]interface{}{
		"parameters": details.Parameters,
	}

	w.WriteHeader(http.StatusOK)
	return respondWithJSON(w, response)
}

func brokerUpdateInstance(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}
	log.Printf("updating service instance %s for broker %s", guids.serviceInstanceGUID, guids.brokerGUID)

	if config.UpdateResponse != 0 {
		w.WriteHeader(config.UpdateResponse)
		return nil
	}

	_, err = store.RetrieveInstance(guids.brokerGUID, guids.serviceInstanceGUID)
	if err != nil {
		return notFoundError{}
	}

	var details resources.ServiceInstanceDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		return newBadRequestError("invalid JSON", err)
	}

	if err := store.UpdateInstance(guids.brokerGUID, guids.serviceInstanceGUID, details); err != nil {
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

func brokerDeprovisionInstance(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}
	log.Printf("deprovisioning service instance %s for broker %s", guids.serviceInstanceGUID, guids.brokerGUID)

	if config.DeprovisionResponse != 0 {
		w.WriteHeader(config.DeprovisionResponse)
		return nil
	}

	if err := store.DeleteInstance(guids.brokerGUID, guids.serviceInstanceGUID); err != nil {
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

func brokerBind(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}
	log.Printf("creating binding %s for service instance %s for broker %s", guids.bindingGUID, guids.serviceInstanceGUID, guids.brokerGUID)

	if config.BindResponse != 0 {
		w.WriteHeader(config.BindResponse)
		return nil
	}

	_, err = store.RetrieveInstance(guids.brokerGUID, guids.serviceInstanceGUID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	var details resources.BindingDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		return newBadRequestError("invalid JSON", err)
	}

	if err := store.CreateBinding(guids.brokerGUID, guids.serviceInstanceGUID, guids.bindingGUID, details); err != nil {
		return err
	}

	response := resources.JSONObject{
		"credentials": resources.JSONObject{
			"username": config.Username,
			"password": config.Password,
		},
	}

	switch config.AsyncResponseDelay {
	case 0:
		w.WriteHeader(http.StatusCreated)
		return respondWithJSON(w, response)
	default:
		return brokerAsyncResponse(w, r, config.AsyncResponseDelay, nil)
	}
}

func brokerGetBinding(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}
	log.Printf("retrieving binding %s for service instance %s for broker %s", guids.bindingGUID, guids.serviceInstanceGUID, guids.brokerGUID)

	if config.GetBindingResponse != 0 {
		w.WriteHeader(config.GetBindingResponse)
		return nil
	}

	details, err := store.RetrieveBinding(guids.brokerGUID, guids.serviceInstanceGUID, guids.bindingGUID)
	if err != nil {
		return err
	}

	details.Credentials = resources.JSONObject{
		"username": config.Username,
		"password": config.Password,
	}

	return respondWithJSON(w, details)
}

func brokerUnbind(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	config, guids, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}
	log.Printf("deleting binding %s for service instance %s for broker %s", guids.bindingGUID, guids.serviceInstanceGUID, guids.brokerGUID)

	if config.UnbindResponse != 0 {
		w.WriteHeader(config.UnbindResponse)
		return nil
	}

	if err := store.DeleteBinding(guids.brokerGUID, guids.serviceInstanceGUID, guids.bindingGUID); err != nil {
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

func brokerLastOperation(store *store.Store, w http.ResponseWriter, r *http.Request) error {
	_, _, err := brokerParseHeaders(store, r)
	if err != nil {
		return err
	}

	var when time.Time
	if err := when.UnmarshalJSON([]byte(`"` + r.FormValue("operation") + `"`)); err != nil {
		return err
	}
	log.Printf("providing last operation status for: %s", when)

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
