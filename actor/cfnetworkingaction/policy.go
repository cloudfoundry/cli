package cfnetworkingaction

import (
	"code.cloudfoundry.org/cli/api/cfnetworking/cfnetv1"
)

type Policy struct {
	SourceName      string
	DestinationName string
	Protocol        string
	StartPort       int
	EndPort         int
}

const policyNotFoundWarning = "Policy does not exist."

func (actor Actor) AllowNetworkAccess(spaceGUID, srcAppName, destAppName, protocol string, startPort, endPort int) (Warnings, error) {
	var allWarnings Warnings

	srcApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(srcAppName, spaceGUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return allWarnings, err
	}

	destApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(destAppName, spaceGUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return allWarnings, err
	}

	err = actor.NetworkingClient.CreatePolicies([]cfnetv1.Policy{
		{
			Source: cfnetv1.PolicySource{
				ID: srcApp.GUID,
			},
			Destination: cfnetv1.PolicyDestination{
				ID:       destApp.GUID,
				Protocol: cfnetv1.PolicyProtocol(protocol),
				Ports: cfnetv1.Ports{
					Start: startPort,
					End:   endPort,
				},
			},
		},
	})
	return allWarnings, err
}

func (actor Actor) ListNetworkAccess(spaceGUID string, srcAppName string) ([]Policy, Warnings, error) {
	var allWarnings Warnings
	var appGUID string

	applications, warnings, err := actor.V3Actor.GetApplicationsBySpace(spaceGUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	if srcAppName != "" {
		srcApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(srcAppName, spaceGUID)
		allWarnings = append(allWarnings, Warnings(warnings)...)
		if err != nil {
			return []Policy{}, allWarnings, err
		}

		appGUID = srcApp.GUID
	}

	var policies []Policy
	v1Policies, err := actor.NetworkingClient.ListPolicies(appGUID)
	if err != nil {
		return []Policy{}, allWarnings, err
	}

	for _, v1Policy := range v1Policies {
		if appGUID == "" || v1Policy.Source.ID == appGUID {
			srcName := ""
			dstName := ""
			for _, app := range applications {
				if v1Policy.Source.ID == app.GUID {
					srcName = app.Name
				}
				if v1Policy.Destination.ID == app.GUID {
					dstName = app.Name
				}
			}
			if srcName != "" && dstName != "" {
				policies = append(policies, Policy{
					SourceName:      srcName,
					DestinationName: dstName,
					Protocol:        string(v1Policy.Destination.Protocol),
					StartPort:       v1Policy.Destination.Ports.Start,
					EndPort:         v1Policy.Destination.Ports.End,
				})
			}
		}
	}

	return policies, allWarnings, nil
}

func (actor Actor) RemoveNetworkAccess(spaceGUID, srcAppName, destAppName, protocol string, startPort, endPort int) (Warnings, error) {
	var allWarnings Warnings

	srcApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(srcAppName, spaceGUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return allWarnings, err
	}

	destApp, warnings, err := actor.V3Actor.GetApplicationByNameAndSpace(destAppName, spaceGUID)
	allWarnings = append(allWarnings, Warnings(warnings)...)
	if err != nil {
		return allWarnings, err
	}

	policyToRemove := cfnetv1.Policy{
		Source: cfnetv1.PolicySource{
			ID: srcApp.GUID,
		},
		Destination: cfnetv1.PolicyDestination{
			ID:       destApp.GUID,
			Protocol: cfnetv1.PolicyProtocol(protocol),
			Ports: cfnetv1.Ports{
				Start: startPort,
				End:   endPort,
			},
		},
	}

	v1Policies, err := actor.NetworkingClient.ListPolicies(srcApp.GUID)
	if err != nil {
		return allWarnings, err
	}

	for _, v1Policy := range v1Policies {
		if v1Policy == policyToRemove {
			return allWarnings, actor.NetworkingClient.RemovePolicies([]cfnetv1.Policy{policyToRemove})
		}
	}

	return append(allWarnings, Warnings([]string{policyNotFoundWarning})...), nil
}
