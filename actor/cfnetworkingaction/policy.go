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

	return policies, allWarnings, err
}
