package formatters

import "strconv"

func InstanceMemoryLimit(limit int64) string {
	if limit == -1 {
		return "Unlimited"
	}

	return strconv.FormatInt(limit, 10) + "M"
}
