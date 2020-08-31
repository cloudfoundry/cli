package shared

import (
	"path"
	"reflect"
	"strconv"
	"strings"

	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/resources"
	"github.com/pkg/errors"
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

func (display *ManifestDiffDisplayer) DisplayDiff(rawManifest []byte, diff resources.ManifestDiff) error {
	// If there are no diffs, just print the manifest
	// TODO: Should we just print something like "no diffs here"?
	if len(diff.Diffs) == 0 {
		display.UI.DisplayDiffUnchanged(string(rawManifest), 0)
		return nil
	}

	//  We unmarshal into a MapSlice vs a map[string]interface{} here to preserve the order of map keys
	var yamlManifest yaml.MapSlice
	err := yaml.Unmarshal(rawManifest, &yamlManifest)
	if err != nil {
		return errors.New("Unable to process manifest diff because its format is invalid.")
	}

	// Distinguish between add/remove and replace diffs
	// Replace diffs have the key in the given manifest so we can navigate to that path and display the changed value
	// Add/Remove diffs will not, so we need to know their parent's path to insert the diff as a child
	pathAddReplaceMap := make(map[string]resources.Diff)
	pathRemoveMap := make(map[string]resources.Diff)
	for _, diff := range diff.Diffs {
		switch diff.Op {
		case resources.AddOperation, resources.ReplaceOperation:
			pathAddReplaceMap[diff.Path] = diff
		case resources.RemoveOperation:
			pathRemoveMap[path.Dir(diff.Path)] = diff
		}
	}

	// Always display the yaml header line
	display.UI.DisplayDiffUnchanged("---", 0)

	// For each entry in the provided manifest, process any diffs at or below that entry
	for _, entry := range yamlManifest {
		display.processDiffsRecursively("/"+interfaceToString(entry.Key), entry.Value, 0, &pathAddReplaceMap, &pathRemoveMap)
	}

	return nil
}

func (display *ManifestDiffDisplayer) processDiffsRecursively(
	currentManifestPath string,
	value interface{},
	depth int,
	pathAddReplaceMap, pathRemoveMap *map[string]resources.Diff,
) {
	field := path.Base(currentManifestPath)

	// If there is a diff at the current path, print it
	if diff, ok := diffExistsAtTheCurrentPath(currentManifestPath, pathAddReplaceMap); ok {
		display.formatDiff(field, diff, depth)
		return
	}

	// If the value is a slice type (i.e. a yaml.MapSlice or a slice), recurse into it
	if isSliceType(value) {
		if isInt(field) {
			// Do not print the numeric values in the diffs/paths used to indicate array position, i.e. /applications/0/env
			display.UI.DisplayDiffUnchanged("-", depth)
		} else {
			display.UI.DisplayDiffUnchanged(field+":", depth)
		}

		if mapSlice, ok := value.(yaml.MapSlice); ok {
			// If a map, recursively process each entry
			for _, entry := range mapSlice {
				display.processDiffsRecursively(
					currentManifestPath+"/"+interfaceToString(entry.Key),
					entry.Value,
					depth+1,
					pathAddReplaceMap, pathRemoveMap,
				)
			}
		} else if asSlice, ok := value.([]interface{}); ok {
			// If a slice, recursively process each element
			for index, sliceValue := range asSlice {
				display.processDiffsRecursively(
					currentManifestPath+"/"+strconv.Itoa(index),
					sliceValue,
					depth+1,
					pathAddReplaceMap, pathRemoveMap,
				)
			}
		}

		// Print add/remove diffs after printing the rest of the map or slice
		if diff, ok := diffExistsAtTheCurrentPath(currentManifestPath, pathRemoveMap); ok {
			display.formatDiff(path.Base(diff.Path), diff, depth+1)
		}

		return
	}

	// Otherwise, print the unchanged field and value
	display.UI.DisplayDiffUnchanged(formatKeyValue(field, value), depth)
}

func (display *ManifestDiffDisplayer) formatDiff(field string, diff resources.Diff, depth int) {
	switch diff.Op {
	case resources.AddOperation:
		display.UI.DisplayDiffAddition(formatKeyValue(field, diff.Value), depth)

	case resources.ReplaceOperation:
		display.UI.DisplayDiffRemoval(formatKeyValue(field, diff.Was), depth)
		display.UI.DisplayDiffAddition(formatKeyValue(field, diff.Value), depth)

	case resources.RemoveOperation:
		display.UI.DisplayDiffRemoval(formatKeyValue(field, diff.Was), depth)
	}
}

func diffExistsAtTheCurrentPath(currentManifestPath string, pathDiffMap *map[string]resources.Diff) (resources.Diff, bool) {
	diff, ok := (*pathDiffMap)[currentManifestPath]
	return diff, ok
}

func formatKeyValue(key string, value interface{}) string {
	if isInt(key) {
		return "-\n" + indentOneLevelDeeper(interfaceToString(value))
	}

	if isMapType(value) || isSliceType(value) {
		return key + ":\n" + indentOneLevelDeeper(interfaceToString(value))
	}

	return key + ": " + interfaceToString(value)
}

func interfaceToString(value interface{}) string {
	val, _ := yaml.Marshal(value)
	return strings.TrimSpace(string(val))
}

func indentOneLevelDeeper(input string) string {
	inputLines := strings.Split(input, "\n")
	outputLines := make([]string, len(inputLines))

	for i, line := range inputLines {
		outputLines[i] = "  " + line
	}

	return strings.Join(outputLines, "\n")
}

func isSliceType(value interface{}) bool {
	valueType := reflect.TypeOf(value)
	return valueType != nil && valueType.Kind() == reflect.Slice
}

func isMapType(value interface{}) bool {
	valueType := reflect.TypeOf(value)
	return valueType != nil && valueType.Kind() == reflect.Map
}

func isInt(field string) bool {
	_, err := strconv.Atoi(field)
	return err == nil
}
