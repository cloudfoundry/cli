package app

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"code.cloudfoundry.org/cli/v8/integration/assets/hydrabroker/store"
	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
)

func respondWithJSON(w http.ResponseWriter, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	log.Printf("responding with JSON: %s", string(bytes))
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	return err
}

func readGUIDs(r *http.Request) (requestGUIDs, error) {
	vars := mux.Vars(r)
	brokerGUID, ok := vars["broker_guid"]
	if !ok {
		return requestGUIDs{}, errors.New("no brokerGUID in request")
	}

	instanceGUID := vars["instance_guid"]
	bindingGUID := vars["binding_guid"]

	return requestGUIDs{
		brokerGUID:          store.BrokerID(brokerGUID),
		serviceInstanceGUID: store.InstanceID(instanceGUID),
		bindingGUID:         store.BindingID(bindingGUID),
	}, nil
}

func mustGUID() string {
	rawGUID, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return rawGUID.String()
}
