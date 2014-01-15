package generic

func IsSliceable(value interface{}) bool {
	switch value.(type) {
	case []string:
		return true
	case []interface{}:
		return true
	default:
		return false
	}
}
