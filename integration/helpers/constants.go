package helpers

const (
	GUIDRegex                       = "[\\da-f]{8}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{12}"
	ISO8601Regex                    = "\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{1,3}[+-]\\d{4}"
	StaticfileBuildpackStartCommand = "$HOME/boot.sh"
	UserFriendlyDateRegex           = "[A-Z][a-z]{2} \\d{2} [A-Z][a-z]{2} \\d{2}:\\d{2}:\\d{2} [A-Z]+ \\d{4}"
)
