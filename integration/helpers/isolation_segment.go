package helpers

func IsolationSegmentName(name ...string) string {
	if len(name) > 0 {
		return PrefixedRandomName("INTEGRATION-ISOLATION-SEGMENT-" + name[0])
	}
	return PrefixedRandomName("INTEGRATION-ISOLATION-SEGMENT")
}
