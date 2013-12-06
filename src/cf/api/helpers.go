package api

import (
	"cf"
	"fmt"
)

func stringOrNull(s string) string {
	if s == "" {
		return "null"
	}

	return fmt.Sprintf(`%s`, s)
}

func mapToJsonValues(params cf.ParamMap) (vals []string) {
	for key, val := range params {
		switch val := val.(type) {
		case string:
			if val != "null" {
				val = fmt.Sprintf(`"%s"`, val)
			}
			vals = append(vals, fmt.Sprintf(`"%s":%s`, key, val))
		case int:
			vals = append(vals, fmt.Sprintf(`"%s":%d`, key, val))
		case uint64:
			vals = append(vals, fmt.Sprintf(`"%s":%d`, key, val))
		default:
			vals = append(vals, fmt.Sprintf(`"%s":%s`, key, val))
		}
	}
	return
}
