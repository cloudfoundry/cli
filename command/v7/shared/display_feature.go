package shared

func FlagBoolToString(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
