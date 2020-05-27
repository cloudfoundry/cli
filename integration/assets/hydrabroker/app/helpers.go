package app

import (
	"encoding/json"
	"errors"
	"net/http"

	uuid "github.com/nu7hatch/gouuid"

	"github.com/gorilla/mux"
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

func readGUID(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	guid, ok := vars["guid"]
	if !ok {
		return "", errors.New("no guid in request")
	}

	return guid, nil
}

func mustGUID() string {
	rawGUID, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return rawGUID.String()
}
