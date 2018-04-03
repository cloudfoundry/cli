package registry

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"regexp"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
)

type instanceHandler struct {
	username string
	password string
	registry Registry
	logger   boshlog.Logger
	logTag   string
}

func newInstanceHandler(
	username string,
	password string,
	registry Registry,
	logger boshlog.Logger,
) *instanceHandler {
	return &instanceHandler{
		username: username,
		password: password,
		registry: registry,
		logger:   logger,
		logTag:   "registryInstanceHandler",
	}
}

type SettingsResponse struct {
	Settings string `json:"settings"`
	Status   string `json:"status"`
}

func (h *instanceHandler) HandleFunc(w http.ResponseWriter, req *http.Request) {
	h.logger.Debug(h.logTag, "Received %s %s", req.Method, req.URL.Path)
	instanceID, ok := h.getInstanceID(req)
	if !ok {
		h.logger.Debug(h.logTag, "Instance ID not found in request:", req.Method)
		h.handleNotFound(w)
		return
	}

	h.logger.Debug(h.logTag, "Found instance ID in request: %s", instanceID)

	switch req.Method {
	case "GET":
		h.HandleGet(instanceID, w, req)
		return
	case "PUT":
		h.HandlePut(instanceID, w, req)
		return
	case "DELETE":
		h.HandleDelete(instanceID, w, req)
		return
	default:
		h.handleNotFound(w)
		return
	}
}

func (h *instanceHandler) HandleGet(instanceID string, w http.ResponseWriter, req *http.Request) {
	settingsJSON, ok := h.registry.Get(instanceID)
	if !ok {
		h.logger.Debug(h.logTag, "No settings for %s found", instanceID)
		h.handleNotFound(w)
		return
	}

	h.logger.Debug(h.logTag, "Found settings for instance %s: %s", instanceID, string(settingsJSON))

	response := SettingsResponse{
		Settings: string(settingsJSON),
		Status:   "ok",
	}

	responseJSON, err := json.Marshal(response)
	if err != nil {
		h.handleBadRequest(w)
		return
	}

	_, err = w.Write(responseJSON)
	if err != nil {
		h.logger.Warn(h.logTag, "Couldn't write response: %s", err.Error())
	}

	return
}

func (h *instanceHandler) HandlePut(instanceID string, w http.ResponseWriter, req *http.Request) {
	if !h.isAuthorized(req) {
		h.handleUnauthorized(w)
		return
	}

	reqBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		h.handleBadRequest(w)
		return
	}

	h.logger.Debug(h.logTag, "Saving settings to registry for instance %s: %s", instanceID, string(reqBody))

	isUpdated := h.registry.Save(instanceID, reqBody)
	if isUpdated {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *instanceHandler) HandleDelete(instanceID string, w http.ResponseWriter, req *http.Request) {
	if !h.isAuthorized(req) {
		h.handleUnauthorized(w)
		return
	}

	h.logger.Debug(h.logTag, "Deleting settings for instance %s", instanceID)
	h.registry.Delete(instanceID)
}

func (h *instanceHandler) handleUnauthorized(w http.ResponseWriter) {
	h.logger.Debug(h.logTag, "Received unauthorized request")
	w.Header().Add("WWW-Authenticate", `Basic realm="Bosh Registry"`)
	w.WriteHeader(http.StatusUnauthorized)
}

func (h *instanceHandler) getInstanceID(req *http.Request) (string, bool) {
	re := regexp.MustCompile("/instances/([^/]+)/settings")
	matches := re.FindStringSubmatch(req.URL.Path)

	if len(matches) == 0 {
		return "", false
	}

	return matches[1], true
}

func (h *instanceHandler) isAuthorized(req *http.Request) bool {
	auth := h.username + ":" + h.password
	expectedAuthorizationHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	return expectedAuthorizationHeader == req.Header.Get("Authorization")
}

func (h *instanceHandler) handleNotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	settingsJSON, err := json.Marshal(SettingsResponse{Status: "not_found"})
	if err != nil {
		h.logger.Warn(h.logTag, "Failed to marshal 'not found' settings response %s", err.Error())
		return
	}
	_, err = w.Write(settingsJSON)
	if err != nil {
		h.logger.Warn(h.logTag, "Failed to write response: %s", err.Error())
	}
}

func (h *instanceHandler) handleBadRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
	settingsJSON, err := json.Marshal(SettingsResponse{Status: "error"})
	if err != nil {
		h.logger.Warn(h.logTag, "Failed to marshal 'bad request' settings response %s", err.Error())
		return
	}
	_, err = w.Write(settingsJSON)
	if err != nil {
		h.logger.Warn(h.logTag, "Failed to write response: %s", err.Error())
	}
}
