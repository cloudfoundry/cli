package helpers

const DefaultTCPRouterGroup = "default-tcp" // Allows for ports 1024-1123

var previouslyUsedPort int

func RandomPort() int {
	if previouslyUsedPort == 0 {
		previouslyUsedPort = 1024
		return previouslyUsedPort
	}

	previouslyUsedPort++
	if previouslyUsedPort > 1123 {
		panic("all ports used, figure out how to fix this future us")
	}

	return previouslyUsedPort
}
