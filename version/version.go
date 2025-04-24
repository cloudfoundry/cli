package version

import (
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver/v4"
)

const DefaultVersion = "0.0.0-unknown-version"

var (
	binaryVersion   string
	binarySHA       string
	binaryBuildDate string
	customVersion   string

	ExitFunc = os.Exit
)

func SetVersion(version string) {
	if version == "" {
		customVersion = ""
		return
	}

	_, err := semver.Parse(version)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid semantic version format: %s\n", err)
		ExitFunc(1)
	}

	customVersion = version
}

func VersionString() string {
	if customVersion != "" {
		return customVersion
	}

	// Remove the "v" prefix from the binary in case it is present
	binaryVersion = strings.TrimPrefix(binaryVersion, "v")
	versionString, err := semver.Make(binaryVersion)
	if err != nil {
		versionString = semver.MustParse(DefaultVersion)
	}

	metaData := []string{}
	if binarySHA != "" {
		metaData = append(metaData, binarySHA)
	}

	if binaryBuildDate != "" {
		metaData = append(metaData, binaryBuildDate)
	}

	versionString.Build = metaData

	return versionString.String()
}
