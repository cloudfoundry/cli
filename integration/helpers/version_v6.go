// +build !V7

package helpers

// SkipIfV7AndVersionLessThan is used to skip tests if the target build is V7 and API version < the specified version
// If minVersion contains the prefix 3 then the v3 version is checked, otherwise the v2 version is used.
func SkipIfV7AndVersionLessThan(minVersion string) {}

// SkipIfV7 is used to skip tests if the target build is V7.
func SkipIfV7() {}
