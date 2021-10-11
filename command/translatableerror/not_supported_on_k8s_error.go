package translatableerror

// NotSupportedOnKubernetesArgumentError represents an error caused by using an
// argument that is not supported on kubernetes
type NotSupportedOnKubernetesArgumentError struct {
	Arg string
}

func (NotSupportedOnKubernetesArgumentError) DisplayUsage() {}

func (NotSupportedOnKubernetesArgumentError) Error() string {
	return "The argument {{.Arg}} is not supported on Kubernetes"
}

func (e NotSupportedOnKubernetesArgumentError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error(), map[string]interface{}{
		"Arg": e.Arg,
	})
}
