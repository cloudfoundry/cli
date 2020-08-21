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
	var yamlManifest yaml.MapSlice
	err := yaml.Unmarshal(rawManifest, &yamlManifest)
	if err != nil {
		log.Errorln("unmarshaling manifest for diff:", err)
	}

	pathDiffMap := make(map[string]resources.Diff)
	pathRemoveMap := make(map[string]resources.Diff)
	for _, diff := range diff.Diffs {
		switch diff.Op {
		case resources.AddOperation, resources.ReplaceOperation:
			pathDiffMap[diff.Path] = diff
		case resources.RemoveOperation:
			pathRemoveMap[path.Dir(diff.Path)] = diff
		}
	}

	// todo: will this ever run more than once?
	display.UI.DisplayDiffUnchanged("---", 0)
	for _, entry := range yamlManifest {
		display.generateDiff("/"+interfaceToString(entry.Key), entry.Value, 0, &pathDiffMap, &pathRemoveMap)
	}
}

func (display *ManifestDiffDisplayer) generateDiff(currentManifestPath string, value interface{}, depth int, pathDiffMap, pathRemoveMap *map[string]resources.Diff) {
	field := path.Base(currentManifestPath)

	if diff, ok := (*pathDiffMap)[currentManifestPath]; ok {
		display.displayDiff(field, diff, depth)
	} else {

		switch reflect.TypeOf(value).Kind() {
		case reflect.String, reflect.Bool, reflect.Int:
			display.UI.DisplayDiffUnchanged(formatKeyValue(field, value), depth)
		case reflect.Slice:
			if _, err := strconv.Atoi(field); err != nil {
				display.UI.DisplayDiffUnchanged(field+":", depth)
			}

			if asSlice, isSlice := value.([]interface{}); isSlice {
				for index, sliceValue := range asSlice {
					display.generateDiff(currentManifestPath+"/"+strconv.Itoa(index), sliceValue, depth, pathDiffMap, pathRemoveMap)
				}

			} else if mapSlice, ok := value.(yaml.MapSlice); ok {
				if diff, ok := (*pathRemoveMap)[currentManifestPath]; ok {
					display.displayDiff(path.Base(diff.Path), diff, depth+1)
				}
				for _, entry := range mapSlice {
					display.generateDiff(currentManifestPath+"/"+interfaceToString(entry.Key), entry.Value, depth+1, pathDiffMap, pathRemoveMap)
				}
			}

		default:
			fmt.Printf("\nunknown kind: %+v", reflect.TypeOf(value).Kind())
			fmt.Printf("\n   for value: %+v\n", value)
			// TODO what do to here? maybe a warning?
			//panic("unknown kind?")
		}
	}
}

func (display *ManifestDiffDisplayer) displayDiff(field string, diff resources.Diff, depth int) {
	switch diff.Op {
	case resources.AddOperation:
		// fmt.Printf("%T\n", diff.Value)
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
