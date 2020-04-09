package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"

	validator "github.com/go-playground/validator/v10"
)

func configCreateBroker(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	var c config.BrokerConfiguration
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		return err
	}

	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}

	for si := range c.Services {
		c.Services[si].GUID = mustGUID()
		c.Services[si].Description = fmt.Sprintf("Description for service offering %s", c.Services[si].Name)
		for pi := range c.Services[si].Plans {
			c.Services[si].Plans[pi].GUID = mustGUID()
			c.Services[si].Plans[pi].Description = fmt.Sprintf("Description for service plan %s", c.Services[si].Plans[pi].Name)
		}
	}

	guid, err := store.CreateBroker(c)
	if err != nil {
		return err
	}

	w.WriteHeader(http.StatusCreated)
	return respondWithJSON(w, config.NewBrokerResponse{GUID: guid})
}

func configDeleteBroker(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	guid, err := readGUID(r)
	if err != nil {
		return err
	}

	store.DeleteBroker(guid)

	w.WriteHeader(http.StatusNoContent)
	return nil
}

func configListBrokers(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	guids := store.ListBrokers()
	return respondWithJSON(w, guids)
}
