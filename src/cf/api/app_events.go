package api

import (
	"cf/configuration"
	"cf/models"
	"cf/net"
	"fmt"
	"generic"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type AppEventsRepository interface {
	ListEvents(appGuid string, cb func([]models.EventFields) bool) net.ApiResponse
}

type CloudControllerAppEventsRepository struct {
	config  *configuration.Configuration
	gateway net.Gateway
}

func NewCloudControllerAppEventsRepository(config *configuration.Configuration, gateway net.Gateway) (repo CloudControllerAppEventsRepository) {
	repo.config = config
	repo.gateway = gateway
	return
}

func (repo CloudControllerAppEventsRepository) ListEvents(appGuid string, cb func([]models.EventFields) bool) net.ApiResponse {
	apiResponse := repo.gateway.ListPaginatedResources(
		repo.config.Target,
		repo.config.AccessToken,
		fmt.Sprintf("/v2/events?q=%s", url.QueryEscape(fmt.Sprintf("actee:%s", appGuid))),
		EventResourceNewV2{},
		func(m []interface{}) bool {
			return cb(convertEvents(m))
		})

	if apiResponse.IsNotFound() {
		apiResponse = repo.gateway.ListPaginatedResources(
			repo.config.Target,
			repo.config.AccessToken,
			fmt.Sprintf("/v2/apps/%s/events", appGuid),
			EventResourceOldV2{},
			func(m []interface{}) bool {
				return cb(convertEvents(m))
			})
	}

	return apiResponse
}

func convertEvents(modelSlice []interface{}) []models.EventFields {
	events := make([]models.EventFields, 0, len(modelSlice))
	for _, model := range modelSlice {
		events = append(events, model.(models.EventFields))
	}
	return events
}

const APP_EVENT_TIMESTAMP_FORMAT = "2006-01-02T15:04:05-07:00"

type EventResourceOldV2 struct {
	Resource
	Entity struct {
		Timestamp       time.Time
		ExitDescription string `json:"exit_description"`
		ExitStatus      int    `json:"exit_status"`
		InstanceIndex   int    `json:"instance_index"`
	}
}

func (resource EventResourceOldV2) ToFields() interface{} {
	return models.EventFields{
		Guid:        resource.Metadata.Guid,
		Name:        "app crashed",
		Timestamp:   resource.Entity.Timestamp,
		Description: fmt.Sprintf("instance: %d, reason: %s, exit_status: %s", resource.Entity.InstanceIndex, resource.Entity.ExitDescription, strconv.Itoa(resource.Entity.ExitStatus)),
	}
}

type EventResourceNewV2 struct {
	Resource
	Entity struct {
		Timestamp time.Time
		Type      string
		Metadata  map[string]interface{}
	}
}

var KNOWN_METADATA_KEYS = []string{
	"index",
	"reason",
	"exit_description",
	"exit_status",
	"recursive",
	"disk_quota",
	"instances",
	"memory",
	"state",
	"command",
	"environment_json",
}

func (resource EventResourceNewV2) ToFields() interface{} {
	metadata := generic.NewMap(resource.Entity.Metadata)
	if metadata.Has("request") {
		metadata = generic.NewMap(metadata.Get("request"))
	}

	return models.EventFields{
		Guid:        resource.Metadata.Guid,
		Name:        resource.Entity.Type,
		Timestamp:   resource.Entity.Timestamp,
		Description: formatDescription(metadata, KNOWN_METADATA_KEYS),
	}
}

func formatDescription(metadata generic.Map, keys []string) string {
	parts := []string{}
	for _, key := range keys {
		value := metadata.Get(key)
		if value != nil {
			parts = append(parts, fmt.Sprintf("%s: %s", key, formatDescriptionPart(value)))
		}
	}
	return strings.Join(parts, ", ")
}

func formatDescriptionPart(val interface{}) string {
	switch val := val.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, byte('f'), -1, 64)
	case bool:
		if val {
			return "true"
		} else {
			return "false"
		}
	default:
		return fmt.Sprintf("%s", val)
	}
}
