package v7action

import (
	"fmt"
	// "regexp"
	// "strings"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	jsonpatch "github.com/evanphx/json-patch"
)

type ManifestDiff struct {
	op    string
	path  string
	was   string
	value string
}

func (actor Actor) SetSpaceManifest(spaceGUID string, rawManifest []byte) (Warnings, error) {
	var allWarnings Warnings
	jobURL, applyManifestWarnings, err := actor.CloudControllerClient.UpdateSpaceApplyManifest(
		spaceGUID,
		rawManifest,
	)
	allWarnings = append(allWarnings, applyManifestWarnings...)
	if err != nil {
		return allWarnings, err
	}

	pollWarnings, err := actor.CloudControllerClient.PollJob(jobURL)
	allWarnings = append(allWarnings, pollWarnings...)
	if err != nil {
		if newErr, ok := err.(ccerror.V3JobFailedError); ok {
			return allWarnings, actionerror.ApplicationManifestError{Message: newErr.Detail}
		}
		return allWarnings, err
	}
	return allWarnings, nil
}

func (actor Actor) GetSpaceManifestDiff(spaceGUID string, newManifest []byte) (string, Warnings, error) {
	manifestDiff, warnings, err := actor.CloudControllerClient.GetSpaceManifestDiff(spaceGUID, newManifest)
	if err != nil {
		return "", Warnings(warnings), err
	}
	fmt.Print("\n\n")
	fmt.Printf("manifestDiff: %v", string(manifestDiff))
	fmt.Print("\n\n")

	fmt.Printf("newManifest: %v", string(newManifest))

	patch, err := jsonpatch.DecodePatch(manifestDiff)
	if err != nil {
		panic(err)
	}

	fmt.Print("\n\n")
	fmt.Printf("patch: %v", patch)

	modified, err := patch.Apply(newManifest)
	if err != nil {
		panic(err)
	}

	fmt.Print("\n\n")
	fmt.Printf("modified: %v", string(modified))

	// path := "/"
	// splitDiff := strings.Split(string(newManifest), "\n")[1:]
	// fmt.Printf("splitDiff: %v", string(splitDiff))
	//for _, split_diff := range splitDiff {
	//	key := regexp.MustCompile(`^\s*-*\s*([[:alnum:]]+):*`).FindStringSubmatch(split_diff)[1]
	//	path += string(key)

	//	val := regexp.MustCompile(`\:(.*)`).FindStringSubmatch(split_diff)[1]
	//	//see if there is a value
	//	fmt.Printf("path: %v \n", path)
	//	fmt.Printf("val: %v \n", val)
	//	if val == "" {
	//		path += "/"
	//	}

	// returnManifest += ("\n" + split_diff)

	// }
	// 	fmt.Print(split_diff + "\n")
	// 	r := regexp.MustCompile(`^\s*-*\s*([[:alnum:]]+):*`)
	// 	fmt.Printf("%#v\n", r.FindStringSubmatch(split_diff)[1])
	// }
	// fmt.Printf("return manifest: %v", returnManifest)
	return "---", Warnings(warnings), err
}

// func fdsffd () {
// 	splitDiff := strings.Split(string(newManifest), "\n")
// 	return_manifest = []
// 	i = 1 // first line ---
// 	path = ""
// 	while i < len(splitDiff)
// 		line = splitdiff[i]
// 		key = regex for thing before :
// 		parent_path = path
// 		path = path + key
// 		value = regex for after :
// 		if value
// 			check if path is in diff
// 				add original line
// 				add line to the return manifest based on diff logic?? //this!! (copy the previous indentation)
// 				path = parent
// 				i = i+1
// 			else (this was not in the diff)
// 				add original line
// 				i = i +1
// 				path = parent
// 		else (the path is not done yet)
// 			add original line
// 			path = path + /
// 			i = i+1

// }
