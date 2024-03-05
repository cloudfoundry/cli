package helpers

import (
	"runtime"

	. "github.com/onsi/ginkgo/v2"
)

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func SkipIfWindows() {

	if IsWindows() {
		Skip("the OS is Windows")
	}

}
