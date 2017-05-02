// +build !windows

package generic

// ExecutableFilename appends does something on Windows, but it is a no-op
// on the many flavors of UNIX.
func ExecutableFilename(name string) string {
	return name
}
