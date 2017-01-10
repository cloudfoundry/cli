package version

import "github.com/blang/semver"

const DefaultVersion = "0.0.0-unknown-version"

var (
	binaryVersion   string
	binarySHA       string
	binaryBuildDate string
)

func VersionString() string {
	versionString, err := semver.Make(binaryVersion)
	if err != nil {
		versionString, _ = semver.Make(DefaultVersion)
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
