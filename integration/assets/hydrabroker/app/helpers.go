package app

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
)

func respondWithJSON(w http.ResponseWriter, data interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

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

	return requestGUIDs{brokerGUID: brokerGUID, serviceInstanceGUID: instanceGUID}, nil
}

func mustGUID() string {
	rawGUID, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return rawGUID.String()
}
