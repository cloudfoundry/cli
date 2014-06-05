package formatters

func Allowed(allowed bool) string {
	if allowed {
		return T("allowed")
	} else {
		return T("disallowed")
	}
}
