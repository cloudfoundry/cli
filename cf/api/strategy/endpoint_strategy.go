package strategy

type EndpointStrategy struct {
	EventsEndpointStrategy
	DomainsEndpointStrategy
}

func NewEndpointStrategy(versionString string) EndpointStrategy {
	version, err := ParseVersion(versionString)
	if err != nil {
		version = Version{0, 0, 0}
	}

	strategy := EndpointStrategy{
		EventsEndpointStrategy:  eventsEndpointStrategy{},
		DomainsEndpointStrategy: domainsEndpointStrategy{},
	}

	if version.GreaterThanOrEqualTo(Version{2, 1, 0}) {
		strategy.EventsEndpointStrategy = globalEventsEndpointStrategy{}
		strategy.DomainsEndpointStrategy = separatedDomainsEndpointStrategy{}
	}

	return strategy
}
