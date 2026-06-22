package ccversion

const (
	MinSupportedV2ClientVersion = "2.128.0"
	MinSupportedClientVersionV9 = "3.160.0"

	MinVersionCreateServiceBrokerV3            = "3.72.0"
	MinVersionCreateSpaceScopedServiceBrokerV3 = "3.75.0"

	MinVersionHTTP2RoutingV3   = "3.104.0"
	MinVersionSpaceSupporterV3 = "3.104.0"

	MinVersionLogRateLimitingV3 = "3.125.0"
	MinVersionPerRouteOpts      = "3.183.0"

	MinVersionBuildpackLifecycleQuery = "3.194.0"

	MinVersionCanarySteps = "3.189.0"

	MinVersionServiceBindingStrategy = "3.205.0"

	MinVersionEmbeddedProcessInstances = "3.211.0"

	MinVersionUpdateStack = "3.211.0"

	// MinVersionRoutePolicies is a placeholder until the CAPI team confirms the
	// version that introduces /v3/route_policies and the enforce_route_policies /
	// route_policies_scope domain fields. Replace "3.999.0" with the real version
	// once known. The test in minimum_version_test.go will keep failing until then.
	MinVersionRoutePolicies = "3.221.0"
)
