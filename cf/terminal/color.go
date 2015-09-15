package terminal

import (
	"fmt"
	"os"
	"regexp"
	"runtime"

	"golang.org/x/crypto/ssh/terminal"
)

type Color uint

const (
	red    Color = 31
	green        = 32
	yellow       = 33
	//	blue          = 34
	magenta = 35
	cyan    = 36
	grey    = 37
	white   = 38
)

var (
	colorize               func(message string, color Color, bold int) string
	OsSupportsColors       = runtime.GOOS != "windows"
	TerminalSupportsColors = isTerminal()
	UserAskedForColors     = ""
)

func init() {
	InitColorSupport()
}

func InitColorSupport() {
	if colorsEnabled() {
		colorize = func(message string, color Color, bold int) string {
			return fmt.Sprintf("\033[%d;%dm%s\033[0m", bold, color, message)
		}
	} else {
		colorize = func(message string, _ Color, _ int) string {
			return message
		}
	}
}

func colorsEnabled() bool {
	return userDidNotDisableColor() &&
		(userEnabledColors() || (TerminalSupportsColors && OsSupportsColors))
}

func userEnabledColors() bool {
	return UserAskedForColors == "true" || os.Getenv("CF_COLOR") == "true"
}

func userDidNotDisableColor() bool {
	return os.Getenv("CF_COLOR") != "false" && (UserAskedForColors != "false" || os.Getenv("CF_COLOR") == "true")
}

func Colorize(message string, color Color) string {
	return colorize(message, color, 0)
}

func ColorizeBold(message string, color Color) string {
	return colorize(message, color, 1)
}

var decolorizerRegex = regexp.MustCompile(`\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[m|K]`)

func Decolorize(message string) string {
	return string(decolorizerRegex.ReplaceAll([]byte(message), []byte("")))
}

func HeaderColor(message string) string {
	return ColorizeBold(message, white)
}

func CommandColor(message string) string {
	return ColorizeBold(message, yellow)
}

func StoppedColor(message string) string {
	return ColorizeBold(message, grey)
}

func AdvisoryColor(message string) string {
	return ColorizeBold(message, yellow)
}

func CrashedColor(message string) string {
	return ColorizeBold(message, red)
}

func FailureColor(message string) string {
	return ColorizeBold(message, red)
}

func SuccessColor(message string) string {
	return ColorizeBold(message, green)
}

func EntityNameColor(message string) string {
	return ColorizeBold(message, cyan)
}

func PromptColor(message string) string {
	return ColorizeBold(message, cyan)
}

func TableContentHeaderColor(message string) string {
	return ColorizeBold(message, cyan)
}

func WarningColor(message string) string {
	return ColorizeBold(message, magenta)
}

func LogStdoutColor(message string) string {
	return Colorize(message, white)
}

func LogStderrColor(message string) string {
	return Colorize(message, red)
}

func LogHealthHeaderColor(message string) string {
	return Colorize(message, grey)
}

func LogAppHeaderColor(message string) string {
	return ColorizeBold(message, yellow)
}

func LogSysHeaderColor(message string) string {
	return ColorizeBold(message, cyan)
}

func isTerminal() bool {
	return terminal.IsTerminal(1)
}
