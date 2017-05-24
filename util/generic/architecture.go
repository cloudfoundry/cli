package generic

func GeneratePlatform(runtimeGOOS string, runtimeGOARCH string) string {
	switch {
	case runtimeGOOS == "linux" && runtimeGOARCH == "amd64":
		return "linux64"
	case runtimeGOOS == "linux" && runtimeGOARCH == "386":
		return "linux32"
	case runtimeGOOS == "windows" && runtimeGOARCH == "amd64":
		return "win64"
	case runtimeGOOS == "windows" && runtimeGOARCH == "386":
		return "win32"
	case runtimeGOOS == "darwin":
		return "osx"
	default:
		return ""
	}
}
