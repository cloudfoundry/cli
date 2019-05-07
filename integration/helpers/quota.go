package helpers

// QuotaName provides a random name prefixed with INTEGRATION-QUOTA. If given a name,
// it structures the name like INTEGRATION-QUOTA-name-randomstring.
func QuotaName(name ...string) string {
	if len(name) > 0 {
		return PrefixedRandomName("INTEGRATION-QUOTA-" + name[0])
	}
	return PrefixedRandomName("INTEGRATION-QUOTA")
}
