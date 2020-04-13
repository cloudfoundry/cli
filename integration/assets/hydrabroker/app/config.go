package app

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/store"
)

var validate = validator.New()

func configCreateBroker(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	c, err := configParse(r.Body)
	if err != nil {
		return err
	}

	guid := store.CreateBroker(c)

	w.WriteHeader(http.StatusCreated)
	return respondWithJSON(w, config.NewBrokerResponse{GUID: guid})
}

func configRecreateBroker(store *store.BrokerConfigurationStore, w http.ResponseWriter, r *http.Request) error {
	c, err := configParse(r.Body)
	if err != nil {
		return err
	}

	guid, err := readGUID(r)
	if err != nil {
		return err
	}

	store.UpdateBroker(guid, c)

	w.WriteHeader(http.StatusNoContent)
	return nil
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

func configParse(body io.ReadCloser) (config.BrokerConfiguration, error) {
	var c config.BrokerConfiguration
	if err := json.NewDecoder(body).Decode(&c); err != nil {
		return config.BrokerConfiguration{}, newBadRequestError("invalid JSON", err)
	}

	if err := validate.Struct(c); err != nil {
		return config.BrokerConfiguration{}, newBadRequestError("invalid body", err)
	}

	fill := func(s string) string {
		if s == "" {
			return mustGUID()
		}
		return s
	}

	for si := range c.Services {
		c.Services[si].ID = fill(c.Services[si].ID)
		c.Services[si].Description = fill(c.Services[si].Description)
		for pi := range c.Services[si].Plans {
			c.Services[si].Plans[pi].ID = fill(c.Services[si].Plans[pi].ID)
			c.Services[si].Plans[pi].Description = fill(c.Services[si].Plans[pi].Description)
		}
	}

	return c, nil
}
