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
	fmt.Printf("DEBUG3: %+v\n", string(rawManifest))
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

		fmt.Printf("DEBUG2: %+v\n", entry)
		fmt.Printf("DEBUG4: %+v\n\n%+v\n\n%+v\n", &diffMap, &removes, entry)

		display.diffValue("/"+interfaceToString(entry.Key), entry.Value, 0, &diffMap, &removes)

		fmt.Printf("DEBUG5: %+v\n\n%+v\n", &diffMap, &removes)
	}
}

// ---
// applications:
//   - name: dora
// 	  old stuff
// 		new thing:
// 		old stuff
// 	- new stuff, type: map
// 		- routes:
// 		  route1
// 			route2

// 	{Op: Add Path: 'applications/0/key' Value: ''}
// 	{Op: Add Path: 'applications/1' Value: []map([]map)}
// ---
// applications:
//  - name: dora
// 	  old stuff
// 		new thing:
// 		old stuff
// 	- new stuff, type: map
// 		- routes:
// 		  route1
// 			route2
// 	- stuff2, type: map
// 		- routes:
// 		  route1
// 			route2

func (display *ManifestDiffDisplayer) diffValue(currentManifestPath string, value interface{}, depth int, diffMap, removes *map[string]resources.Diff) {
	field := getLastPart(currentManifestPath)

	fmt.Printf("DEBUG1: %+v: %+v\n", reflect.TypeOf(value).Kind(), field)

	switch reflect.TypeOf(value).Kind() {
	case reflect.String, reflect.Bool, reflect.Int:
		if diff, ok := (*diffMap)[currentManifestPath]; ok {
			display.displayDiff(field, diff, depth)
		} else {
			display.UI.DisplayDiffUnchanged(formatKeyValue(field, value), depth)
		}
	case reflect.Slice:
		if asMapSlice, isMapSlice := value.([]yaml.MapSlice); isMapSlice {
			fmt.Printf("DEBUG6: asMapSlice: %+v\n", asMapSlice)
			display.UI.DisplayDiffUnchanged(field+":", depth)

			for index, sliceValue := range asMapSlice {
				display.diffValue(currentManifestPath+"/"+strconv.Itoa(index), sliceValue, depth+1, diffMap, removes)
			}
		}

		if asSlice, isSlice := value.([]interface{}); isSlice {
			fmt.Printf("DEBUG7: asSlice: %+v\n", asSlice)
			display.UI.DisplayDiffUnchanged(field+":", depth)

			for index, sliceValue := range asSlice {
				display.diffValue(currentManifestPath+"/"+strconv.Itoa(index), sliceValue, depth+1, diffMap, removes)
			}
		} else {
			fmt.Printf("DEBUG8: in else case")
			if diff, ok := (*diffMap)[currentManifestPath]; ok {
				display.displayDiff(field, diff, depth)
			} else {
				display.UI.DisplayDiffUnchanged(field+":", depth)

				if diff, ok := (*removes)[currentManifestPath]; ok {
					display.displayDiff(path.Base(diff.Path), diff, depth+1)
				}

				mapSlice := value.(yaml.MapSlice)
				for _, entry := range mapSlice {
					display.diffValue(currentManifestPath+"/"+interfaceToString(entry.Key), entry.Value, depth+1, diffMap, removes)
				}
			}
		}

	default:
		fmt.Printf("\nunknown kind: %+v", reflect.TypeOf(value).Kind())
		fmt.Printf("\n   for value: %+v\n", value)
		// TODO what do to here?
		//panic("unknown kind?")
	}
}

func (display *ManifestDiffDisplayer) displayDiff(field string, diff resources.Diff, depth int) {
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

func formatKeyValue(key string, value interface{}) string {
	return key + ": " + interfaceToString(value)
}

func getLastPart(path string) string {
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

func interfaceToString(value interface{}) string {
	return fmt.Sprintf("%v", value)
}
