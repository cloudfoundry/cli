package utils

import (
	"strconv"
	"strings"
)

type Version struct {
	Revisions []int
}

func NewVersion(version string) Version {
	split := strings.Split(version, ".")

	retVersion := Version{[]int{}}
	for _, v := range split {
		parsed, _ := strconv.Atoi(v)
		retVersion.Revisions = append(retVersion.Revisions, parsed)
	}
	return retVersion
}

func (ver Version) GreaterThanOrEqual(target Version) bool {
	if len(ver.Revisions) > 0 && len(target.Revisions) == 0 {
		return true
	} else if len(target.Revisions) > 0 && len(ver.Revisions) == 0 {
		for _, v := range target.Revisions {
			if v != 0 {
				return false
			}
		}
		return true
	}

	if ver.Revisions[0] < target.Revisions[0] {
		return false
	} else if (len(ver.Revisions) > 1 || len(target.Revisions) > 1) && ver.Revisions[0] == target.Revisions[0] {
		return Version{ver.Revisions[1:]}.GreaterThanOrEqual(Version{target.Revisions[1:]})
	}

	return true
}
