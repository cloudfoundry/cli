package uihelpers

import "strings"

func ParseTags(tags string) []string {
	tags = strings.Trim(tags, `"`)
	tagsList := strings.Split(tags, ",")
	finalTagsList := []string{}
	for _, tag := range tagsList {
		trimmed := strings.Trim(tag, " ")
		if trimmed != "" {
			finalTagsList = append(finalTagsList, trimmed)
		}
	}
	return finalTagsList
}
