package models

import (
	"encoding/json"
	"strconv"

	"github.com/cloudfoundry/cli/cf/formatters"
)

type SpaceQuota struct {
	GUID                    string `json:"guid,omitempty"`
	Name                    string `json:"name"`
	MemoryLimit             int64  `json:"memory_limit"`          // in Megabytes
	InstanceMemoryLimit     int64  `json:"instance_memory_limit"` // in Megabytes
	RoutesLimit             int    `json:"total_routes"`
	ServicesLimit           int    `json:"total_services"`
	NonBasicServicesAllowed bool   `json:"non_basic_services_allowed"`
	OrgGUID                 string `json:"organization_guid"`
	AppInstanceLimit        int    `json:"app_instance_limit"`
}

func (q SpaceQuota) FormattedMemoryLimit() string {
	return formatters.ByteSize(q.MemoryLimit * formatters.MEGABYTE)
}

func (q SpaceQuota) FormattedInstanceMemoryLimit() string {
	if q.InstanceMemoryLimit == -1 {
		return "unlimited"
	}
	return formatters.ByteSize(q.InstanceMemoryLimit * formatters.MEGABYTE)
}

func (q SpaceQuota) FormattedAppInstanceLimit() string {
	appInstanceLimit := "unlimited"
	if q.AppInstanceLimit != -1 { //TODO - figure out how to use resources.UnlimitedAppInstances
		appInstanceLimit = strconv.Itoa(q.AppInstanceLimit)
	}

	return appInstanceLimit
}

type SpaceQuotaResponse struct {
	GUID                    string      `json:"guid,omitempty"`
	Name                    string      `json:"name"`
	MemoryLimit             int64       `json:"memory_limit"`          // in Megabytes
	InstanceMemoryLimit     int64       `json:"instance_memory_limit"` // in Megabytes
	RoutesLimit             int         `json:"total_routes"`
	ServicesLimit           int         `json:"total_services"`
	NonBasicServicesAllowed bool        `json:"non_basic_services_allowed"`
	OrgGUID                 string      `json:"organization_guid"`
	AppInstanceLimit        json.Number `json:"app_instance_limit"`
}
