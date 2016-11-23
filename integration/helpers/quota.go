package helpers

func QuotaName(name ...string) string {
	if len(name) > 0 {
		return PrefixedRandomName("INTEGRATION-QUOTA-" + name[0])
	}
	return PrefixedRandomName("INTEGRATION-QUOTA")
}
