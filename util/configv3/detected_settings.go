package configv3

// detectedSettings are automatically detected settings determined by the CLI.
type detectedSettings struct {
	currentDirectory string
	terminalWidth    int
	tty              bool
}
