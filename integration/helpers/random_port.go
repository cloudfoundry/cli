package helpers

var previouslyUsedPort int

// RandomPort returns a port number that has not yet been used, starting at 1024 and
// increasing by one each time it is called. It errors if the number increases above 1123.
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
