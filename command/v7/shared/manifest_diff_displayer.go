package shared

import (
	"fmt"
	"path"
	"reflect"
	"strconv"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/resources"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ManifestDiffDisplayer struct {
	UI command.UI
}

func NewManifestDiffDisplayer(ui command.UI) *ManifestDiffDisplayer {
	return &ManifestDiffDisplayer{
		UI: ui,
	}
}

func (display *ManifestDiffDisplayer) DisplayDiff(rawManifest []byte, diff resources.ManifestDiff) {
	// If there are no diffs, just print the manifest
	// TODO: Should we just print something like "no diffs here"?
	if len(diff.Diffs) == 0 {
		display.UI.DisplayDiffUnchanged(string(rawManifest), 0)
		return
	}

	//  We unmarshal into a MapSlice vs a map[string]{interface} here to preserve the order of map keys
	var yamlManifest yaml.MapSlice
	err := yaml.Unmarshal(rawManifest, &yamlManifest)
	if err != nil {
		log.Errorln("unmarshaling manifest for diff:", err)
	}

	// Distinguish between add/remove and replace diffs
	// Replace diffs have the key in the given manifest so we can navigate to that path and display the changed value
	// Add/Remove diffs will not, so we need to know their parent's path to insert the diff as a child
	pathReplaceMap := make(map[string]resources.Diff)
	pathAddRemoveMap := make(map[string]resources.Diff)
	for _, diff := range diff.Diffs {
		switch diff.Op {
		case resources.AddOperation, resources.ReplaceOperation:
			pathReplaceMap[diff.Path] = diff
		case resources.RemoveOperation:
			pathAddRemoveMap[path.Dir(diff.Path)] = diff
		}
	}

	// Always display the yaml header line
	display.UI.DisplayDiffUnchanged("---", 0)
	// For each entry in the provided manifest, process any diffs at or below that entry
	for _, entry := range yamlManifest {
		display.processDiffsRecursively("/"+interfaceToString(entry.Key), entry.Value, 0, &pathReplaceMap, &pathAddRemoveMap)
	}
}

func (display *ManifestDiffDisplayer) processDiffsRecursively(currentManifestPath string, value interface{}, depth int, pathReplaceMap, pathAddRemoveMap *map[string]resources.Diff) {
	field := path.Base(currentManifestPath)

	// If there is a diff at the current path, print it
	if diff, ok := diffExistsAtTheCurrentPath(currentManifestPath, pathReplaceMap); ok {
		display.formatDiff(field, diff, depth)
		return
	}

	// If the value is a slice type (i.e. a yaml.MapSlice or a slice), recurse into it
	if isSliceType(value) {
		// Do not print the numeric values in the paths used to indicate array position, i.e. /applications/0/env
		if !isInt(field) {
			display.UI.DisplayDiffUnchanged(field+":", depth)
		}

		if mapSlice, ok := value.(yaml.MapSlice); ok {
			// If a map, recursively process each entry
			for _, entry := range mapSlice {
				display.processDiffsRecursively(currentManifestPath+"/"+interfaceToString(entry.Key), entry.Value, depth+1, pathReplaceMap, pathAddRemoveMap)
			}
		} else if asSlice, ok := value.([]interface{}); ok {
			// If a slice, recursively process each element
			for index, sliceValue := range asSlice {
				display.processDiffsRecursively(currentManifestPath+"/"+strconv.Itoa(index), sliceValue, depth, pathReplaceMap, pathAddRemoveMap)
			}
		}

		// Print add/remove diffs after printing the rest of the map or slice
		if diff, ok := (*pathAddRemoveMap)[currentManifestPath]; ok {
			display.formatDiff(path.Base(diff.Path), diff, depth+1)
		}

		return
	}

	// Otherwise, print the unchanged field and value
	display.UI.DisplayDiffUnchanged(formatKeyValue(field, value), depth)
}

func diffExistsAtTheCurrentPath(currentManifestPath string, pathDiffMap *map[string]resources.Diff) (resources.Diff, bool) {
	diff, ok := (*pathDiffMap)[currentManifestPath]
	return diff, ok
}

func isSliceType(value interface{}) bool {
	return reflect.TypeOf(value).Kind() == reflect.Slice
}

func isInt(field string) bool {
	_, err := strconv.Atoi(field)
	return err == nil
}

func (display *ManifestDiffDisplayer) formatDiff(field string, diff resources.Diff, depth int) {
	switch diff.Op {
	case resources.AddOperation:
		if mapDiff, ok := diff.Value.(map[string]interface{}); ok {
			display.UI.DisplayDiffAddition(field+":", depth)
			display.UI.DisplayDiffAdditionForMapStringInterface(mapDiff, depth+1)
		} else if mapDiff, ok := diff.Value.([]interface{}); ok {
			display.UI.DisplayDiffAddition(field+":", depth)
			for _, entry := range mapDiff {
				if mapDiff, ok := convertToMap(entry); ok {
					display.UI.DisplayDiffAdditionForMapStringInterface(mapDiff, depth+1)
				} else if stringDiff, ok := diff.Value.(string); ok {
					display.UI.DisplayDiffAddition(stringDiff, depth)
				}
			}
		} else {
			display.UI.DisplayDiffAddition(formatKeyValue(field, diff.Value), depth)
		}
	case resources.ReplaceOperation:
		display.UI.DisplayDiffRemoval(formatKeyValue(field, diff.Was), depth)
		display.UI.DisplayDiffAddition(formatKeyValue(field, diff.Value), depth)
	case resources.RemoveOperation:
		display.UI.DisplayDiffRemoval(formatKeyValue(field, diff.Was), depth)
	}
}

func convertToMap(value interface{}) (map[string]interface{}, bool) {
	if mapDiff, ok := value.(map[string]interface{}); ok {
		return mapDiff, ok
	}
	return make(map[string]interface{}), false
}

func formatKeyValue(key string, value interface{}) string {
	return key + ": " + interfaceToString(value)
}

func interfaceToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
	// val, _ := yaml.Marshal(value)
	// return string(val)
}
