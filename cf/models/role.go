package models

import (
	"errors"
	"strings"
)

type Role int

const (
	RoleUnknown Role = iota - 1
	RoleOrgUser
	RoleOrgManager
	RoleBillingManager
	RoleOrgAuditor
	RoleSpaceManager
	RoleSpaceDeveloper
	RoleSpaceAuditor
)

var ErrUnknownRole = errors.New("Unknown Role")

func RoleFromString(roleString string) (Role, error) {
	switch strings.ToLower(roleString) {
	case "orgmanager":
		return RoleOrgManager, nil
	case "billingmanager":
		return RoleBillingManager, nil
	case "orgauditor":
		return RoleOrgAuditor, nil
	case "spacemanager":
		return RoleSpaceManager, nil
	case "spacedeveloper":
		return RoleSpaceDeveloper, nil
	case "spaceauditor":
		return RoleSpaceAuditor, nil
	default:
		return RoleUnknown, ErrUnknownRole
	}
}

func (r Role) ToString() string {
	switch r {
	case RoleUnknown:
		return "RoleUnknown"
	case RoleOrgUser:
		return "RoleOrgUser"
	case RoleOrgManager:
		return "RoleOrgManager"
	case RoleBillingManager:
		return "RoleBillingManager"
	case RoleOrgAuditor:
		return "RoleOrgAuditor"
	case RoleSpaceManager:
		return "RoleSpaceManager"
	case RoleSpaceDeveloper:
		return "RoleSpaceDeveloper"
	case RoleSpaceAuditor:
		return "RoleSpaceAuditor"
	default:
		return ""
	}
}

func (r Role) Display() string {
	return strings.TrimPrefix(r.ToString(), "Role")
}
