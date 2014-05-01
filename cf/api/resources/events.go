package resources

import (
	"fmt"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/generic"
	"strconv"
	"strings"
	"time"
)

type EventResource interface {
	ToFields() models.EventFields
}

type EventResourceNewV2 struct {
	Resource
	Entity struct {
		Timestamp time.Time
		Type      string
		ActorName string `json:"actor_name"`
		Metadata  map[string]interface{}
	}
}

type EventResourceOldV2 struct {
	Resource
	Entity struct {
		Timestamp       time.Time
		ExitDescription string `json:"exit_description"`
		ExitStatus      int    `json:"exit_status"`
		InstanceIndex   int    `json:"instance_index"`
	}
}

func (resource EventResourceNewV2) ToFields() models.EventFields {
	metadata := generic.NewMap(resource.Entity.Metadata)
	if metadata.Has("request") {
		metadata = generic.NewMap(metadata.Get("request"))
	}

	return models.EventFields{
		Guid:        resource.Metadata.Guid,
		Name:        resource.Entity.Type,
		Timestamp:   resource.Entity.Timestamp,
		Description: formatDescription(metadata, knownMetadataKeys),
		ActorName:   resource.Entity.ActorName,
	}
}

func (resource EventResourceOldV2) ToFields() models.EventFields {
	return models.EventFields{
		Guid:        resource.Metadata.Guid,
		Name:        "app crashed",
		Timestamp:   resource.Entity.Timestamp,
		Description: fmt.Sprintf("instance: %d, reason: %s, exit_status: %s", resource.Entity.InstanceIndex, resource.Entity.ExitDescription, strconv.Itoa(resource.Entity.ExitStatus)),
	}
}

var knownMetadataKeys = []string{
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
