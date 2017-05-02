// +build windows

package generic

import (
	"fmt"
	"strings"
)

// ExecutableFilename appends '.exe' to a filename when necessary in order to
// make it executable on Windows
func ExecutableFilename(name string) string {
	if strings.HasSuffix(name, ".exe") {
		return name
	}
	return fmt.Sprintf("%s.exe", name)
}
