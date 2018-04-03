package director

import (
	"encoding/json"
	"net/http"
	"strconv"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type ID struct {
	ID string `json:"id"`
}

type DiffConfigBody struct {
	From ID `json:"from"`
	To   ID `json:"to"`
}

func (d DirectorImpl) DiffConfigByID(fromID string, toID string) (ConfigDiff, error) {
	resp, err := d.client.DiffConfigByID(fromID, toID)
	if err != nil {
		return ConfigDiff{}, err
	}
	return NewConfigDiff(resp.Diff), nil
}

func (c Client) DiffConfigByID(fromID string, toID string) (ConfigDiffResponse, error) {
	setHeaders := func(req *http.Request) {
		req.Header.Add("Content-Type", "application/json")
	}

	_, errFromId := strconv.Atoi(fromID)
	_, errToId := strconv.Atoi(toID)

	if (errFromId != nil) || (errToId != nil) {
		return ConfigDiffResponse{}, bosherr.Error("Config ID needs to be an integer.")
	}

	body, err := json.Marshal(DiffConfigBody{ID{fromID}, ID{toID}})
	if err != nil {
		return ConfigDiffResponse{}, bosherr.WrapError(err, "Can't marshal request body")
	}

	return c.postConfigDiff("/configs/diff", body, setHeaders)
}
