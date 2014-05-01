package formatters

func Allowed(allowed bool) string {
	if allowed {
		return "allowed"
	} else {
		return "disallowed"
	}
}
