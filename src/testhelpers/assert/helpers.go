package assert

import (
	"fmt"
	"github.com/onsi/ginkgo"
	"runtime"
	"strings"
)

// Fail reports a failure through
func Fail(failureMessage string, msgAndArgs ...interface{}) bool {

	message := messageFromMsgAndArgs(msgAndArgs...)

	if len(message) > 0 {
		ginkgo.Fail(fmt.Sprintf("\r%s\r\tLocation:\t%s\n\r\tError:\t\t%s\n\r\tMessages:\t%s\n\r", getWhitespaceString(), CallerInfo(), failureMessage, message))
	} else {
		ginkgo.Fail(fmt.Sprintf("\r%s\r\tLocation:\t%s\n\r\tError:\t\t%s\n\r", getWhitespaceString(), CallerInfo(), failureMessage))
	}

	return false
}

/* CallerInfo is necessary because the assert functions use the testing object
internally, causing it to print the file:line of the assert method, rather than where
the problem actually occured in calling code.*/

// CallerInfo returns a string containing the file and line number of the assert call
// that failed.
func CallerInfo() string {

	file := ""
	line := 0
	ok := false

	for i := 0; ; i++ {
		_, file, line, ok = runtime.Caller(i)
		if !ok {
			return ""
		}
		parts := strings.Split(file, "/")
		dir := parts[len(parts)-2]
		file = parts[len(parts)-1]
		if (dir != "assert" && dir != "mock") || file == "mock_test.go" {
			break
		}
	}

	return fmt.Sprintf("%s:%d", file, line)
}

// Takes a slice of errors and asserts that there were none provided
// When failing, appends error messages together on newlines and
// provides a count of how many errors were passed in
func AssertNoErrors(errs []error) {
	if len(errs) > 0 {
		var concatErrors string
		for _, err := range errs {
			concatErrors = concatErrors + err.Error() + "\n"
		}
		ginkgo.Fail(fmt.Sprintf("Expected no errors, but there were %d.\n%s", len(errs), concatErrors))
	}
}

func AssertPanic(panicValue interface{}, callback func()) {
	defer func() {
		value := recover()
		if value != panicValue {
			panic(value)
		}
	}()

	callback()
	ginkgo.Fail("Expected a panic, but got none!")
}

// getWhitespaceString returns a string that is long enough to overwrite the default
// output from the go testing framework.
func getWhitespaceString() string {

	_, file, line, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	parts := strings.Split(file, "/")
	file = parts[len(parts)-1]

	return strings.Repeat(" ", len(fmt.Sprintf("%s:%d:      ", file, line)))

}

func messageFromMsgAndArgs(msgAndArgs ...interface{}) string {
	if len(msgAndArgs) == 0 || msgAndArgs == nil {
		return ""
	}
	if len(msgAndArgs) == 1 {
		return msgAndArgs[0].(string)
	}
	if len(msgAndArgs) > 1 {
		return fmt.Sprintf(msgAndArgs[0].(string), msgAndArgs[1:]...)
	}
	return ""
}
