$VERSION_FILE = $(split-path $MyInvocation.MyCommand.Definition) + "\..\VERSION"
$BUILT_AT_FILE = $(split-path $MyInvocation.MyCommand.Definition) + "\..\BUILT_AT"

$CURRENT_SHA = $(git rev-parse --short HEAD)
$CURRENT_VERSION = get-content VERSION
$VERSION_STRING = $CURRENT_VERSION + "+" + $CURRENT_SHA
$DATE = Get-Date -uformat "%Y-%m-%dT%H:%M:%S+00:00"

$VERSION_STRING | Out-File -Encoding "UTF8" $VERSION_FILE
$DATE | Out-File -Encoding "UTF8" $BUILT_AT_FILE
