package translatableerror

type NetworkPolicyProtocolOrPortNotProvidedError struct{}

func (NetworkPolicyProtocolOrPortNotProvidedError) DisplayUsage() {}

func (NetworkPolicyProtocolOrPortNotProvidedError) Error() string {
	return "Incorrect Usage: --protocol and --port flags must be specified together"
}

func (e NetworkPolicyProtocolOrPortNotProvidedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
