package shared

import (
	"fmt"
	"path"
	"reflect"
	"strconv"
	"strings"

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
	var yamlEntries yaml.MapSlice
	err := yaml.Unmarshal(rawManifest, &yamlEntries)
	if err != nil {
		log.Errorln("unmarshaling manifest for diff:", err)
	}

	diffMap := make(map[string]resources.Diff, len(diff.Diffs))
	removes := make(map[string]resources.Diff)
	for _, diff := range diff.Diffs {
		switch diff.Op {
		case resources.AddOperation, resources.ReplaceOperation:
			diffMap[diff.Path] = diff
		case resources.RemoveOperation:
			removes[path.Dir(diff.Path)] = diff
		}
	}

	display.UI.DisplayDiffUnchanged("---", 0)
	for _, entry := range yamlEntries {
		fmt.Printf("%v\n", entry)
		fmt.Printf("currentManifestPath: /%v\n", interfaceToString(entry.Key))
		fmt.Printf("value: %v\n", entry.Value)
		fmt.Printf("diffMap: %v\n", diffMap)
		fmt.Printf("removes: %v\n", removes)
		display.diffValue("/"+interfaceToString(entry.Key), entry.Value, 0, &diffMap, &removes)
	}
}

func (display *ManifestDiffDisplayer) diffValue(currentManifestPath string, value interface{}, depth int, diffMap, removes *map[string]resources.Diff) {
	field := getLastPart(currentManifestPath)
	fmt.Printf("\nNEW ITERATION!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
	fmt.Printf("currentManifestPath: /%v\n", currentManifestPath)
	fmt.Printf("value: %v\n", value)
	if diff, ok := (*diffMap)[currentManifestPath]; ok {
		display.displayDiff(field, diff, depth)
	} else {

		switch reflect.TypeOf(value).Kind() {
		case reflect.String, reflect.Bool, reflect.Int:
			fmt.Printf("value is a string%v\n", value)
			display.UI.DisplayDiffUnchanged(formatKeyValue(field, value), depth)
		case reflect.Slice:
			// TODO: WE NEVER ENTER THIS CASE
			// if asMapSlice, isMapSlice := value.([]yaml.MapSlice); isMapSlice {
			// 	fmt.Printf("value is a mapslice: %v\n", value)

			// 	display.UI.DisplayDiffUnchanged(field+":", depth)

			// 	for index, sliceValue := range asMapSlice {
			// 		display.diffValue(currentManifestPath+"/"+strconv.Itoa(index), sliceValue, depth+1, diffMap, removes)
			// 	}
			// }
			if asSlice, isSlice := value.([]interface{}); isSlice {
				// TODO: consider adding case here to check for the current manifest path existing in diff map. for additions of routes
				// extra credit: move this case out to be global?
				fmt.Printf("value is a list of lists: %v\n.", value)
				display.UI.DisplayDiffUnchanged(field+":", depth)
				for index, sliceValue := range asSlice {
					display.diffValue(currentManifestPath+"/"+strconv.Itoa(index), sliceValue, depth, diffMap, removes)
				}

			} else {
				fmt.Printf("value is a an else: %v\n.", value)
				if _, err := strconv.Atoi(field); err != nil {
					display.UI.DisplayDiffUnchanged(field+":", depth)
				}

				if diff, ok := (*removes)[currentManifestPath]; ok {
					display.displayDiff(path.Base(diff.Path), diff, depth+1)
				}

				mapSlice := value.(yaml.MapSlice)
				for _, entry := range mapSlice {
					display.diffValue(currentManifestPath+"/"+interfaceToString(entry.Key), entry.Value, depth+1, diffMap, removes)
				}
			}

		default:
			fmt.Printf("\nunknown kind: %+v", reflect.TypeOf(value).Kind())
			fmt.Printf("\n   for value: %+v\n", value)
			// TODO what do to here?
			//panic("unknown kind?")
		}
	}
}

func (display *ManifestDiffDisplayer) displayDiff(field string, diff resources.Diff, depth int) {
	switch diff.Op {
	case resources.AddOperation:
		fmt.Printf("%T\n", diff.Value)
		if mapDiff, ok := diff.Value.(map[string]interface{}); ok {
			display.UI.DisplayDiffAddition(field+":", depth)
			display.UI.DisplayDiffAdditionForMapStringInterface(mapDiff, depth+1)
		} else if mapDiff, ok := diff.Value.([]map[string]interface{}); ok {
			display.UI.DisplayDiffAddition(field+":", depth)
			for _, line := range mapDiff {
				display.UI.DisplayDiffAdditionForMapStringInterface(line, depth+1)
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

func formatKeyValue(key string, value interface{}) string {
	return key + ": " + interfaceToString(value)
}

func interfaceToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
	// val, _ := yaml.Marshal(value)
	// return string(val)
}
