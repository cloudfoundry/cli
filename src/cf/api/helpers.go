package api

import "fmt"

func stringOrNull(s string) string {
	if s == "" {
		return "null"
	}

	return fmt.Sprintf(`"%s"`, s)
}
