package translatableerror

type NetworkPolicyProtocolOrPortNotProvidedError struct{}

func (NetworkPolicyProtocolOrPortNotProvidedError) Error() string {
	return "--protocol and --port flags must be specified together"
}

func (e NetworkPolicyProtocolOrPortNotProvidedError) Translate(translate func(string, ...interface{}) string) string {
	return translate(e.Error())
}
