package shared

import (
	"fmt"
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
	if len(diff.Diffs) == 0 {
		display.UI.DisplayDiffUnchanged(string(rawManifest), 0, false) //Todo this is unnecessary MdL
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
	display.UI.DisplayDiffUnchanged("---", 0, false)

	// For each entry in the provided manifest, process any diffs at or below that entry
	for _, entry := range yamlManifest {
		display.processDiffsRecursively("/"+interfaceToString(entry.Key), entry.Value, 0, &pathAddReplaceMap, &pathRemoveMap, false)
	}

	return nil
}

func (display *ManifestDiffDisplayer) processDiffsRecursively(
	currentManifestPath string,
	value interface{},
	depth int,
	pathAddReplaceMap, pathRemoveMap *map[string]resources.Diff,
	addHyphen bool,
) {
	field := path.Base(currentManifestPath)
	// If there is a diff at the current path, print it
	if diff, ok := diffExistsAtTheCurrentPath(currentManifestPath, pathAddReplaceMap); ok {
		display.formatDiff(field, diff, depth, addHyphen)
		return
	}
	// If the value is a slice type (i.e. a yaml.MapSlice or a slice), recurse into it
	if isSliceType(value) {
		addHyphen = isInt(field)
		if isInt(field) {
		} else {
			display.UI.DisplayDiffUnchanged(field+":", depth, addHyphen)
			fmt.Println("non int case:" + field)
			depth += 1
		}

		if mapSlice, ok := value.(yaml.MapSlice); ok {
			// If a map, recursively process each entry
			for i, entry := range mapSlice {
				if i > 0 {
					addHyphen = false
				}
				display.processDiffsRecursively(
					currentManifestPath+"/"+interfaceToString(entry.Key),
					entry.Value,
					depth,
					pathAddReplaceMap, pathRemoveMap, addHyphen,
				)
			}
		} else if asSlice, ok := value.([]interface{}); ok {
			// If a slice, recursively process each element
			for index, sliceValue := range asSlice {
				display.processDiffsRecursively(
					currentManifestPath+"/"+strconv.Itoa(index),
					sliceValue,
					depth,
					pathAddReplaceMap, pathRemoveMap, false,
				)
			}
		}

		// Print add/remove diffs after printing the rest of the map or slice
		if diff, ok := diffExistsAtTheCurrentPath(currentManifestPath, pathRemoveMap); ok {
			display.formatDiff(path.Base(diff.Path), diff, depth, addHyphen)
		}

		return
	}

	// Otherwise, print the unchanged field and value
	display.UI.DisplayDiffUnchanged(formatKeyValue(field, value), depth, addHyphen)
}

func (display *ManifestDiffDisplayer) formatDiff(field string, diff resources.Diff, depth int, addHyphen bool) {
	addHyphen = isInt(field) || addHyphen
	switch diff.Op {
	case resources.AddOperation:
		display.UI.DisplayDiffAddition(formatKeyValue(field, diff.Value), depth, addHyphen)

	case resources.ReplaceOperation:
		display.UI.DisplayDiffRemoval(formatKeyValue(field, diff.Was), depth, addHyphen)
		display.UI.DisplayDiffAddition(formatKeyValue(field, diff.Value), depth, addHyphen)

	case resources.RemoveOperation:
		display.UI.DisplayDiffRemoval(formatKeyValue(field, diff.Was), depth, addHyphen)
	}
}

func diffExistsAtTheCurrentPath(currentManifestPath string, pathDiffMap *map[string]resources.Diff) (resources.Diff, bool) {
	diff, ok := (*pathDiffMap)[currentManifestPath]
	return diff, ok
}

func formatKeyValue(key string, value interface{}) string {
	if isInt(key) {
		return interfaceToString(value)
	}

	if isMapType(value) || isSliceType(value) {
		return key + ":\n" + interfaceToString(value)
	}

	return key + ": " + interfaceToString(value)
}

func interfaceToString(value interface{}) string {
	val, _ := yaml.Marshal(value)
	return strings.TrimSpace(string(val))
}

// func indentOneLevelDeeper(input string, isInt bool) string {
// 	inputLines := strings.Split(input, "\n")
// 	outputLines := make([]string, len(inputLines))

// 	for i, line := range inputLines {
// 		outputLines[i] = "  " + line
// 	}

// 	return strings.Join(outputLines, "\n")
// }

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
