$APP_CONST_FILE = $(split-path $MyInvocation.MyCommand.Definition) + "\..\cf\app_constants.go"
$APP_CONST_FILE_TMP = $APP_CONST_FILE + ".tmp"
$CURRENT_SHA = $(git rev-parse --short HEAD)
$CURRENT_VERSION = get-content VERSION
$VERSION_STRING = $CURRENT_VERSION + "-" + $CURRENT_SHA
$DATE = Get-Date -uformat "%b %d, %Y %I:%M:%p"

get-content $APP_CONST_FILE | %{$_ -replace "BUILT_FROM_SOURCE", $VERSION_STRING} | Out-File -Encoding "UTF8" $APP_CONST_FILE_TMP
mv -Force $APP_CONST_FILE_TMP $APP_CONST_FILE

get-content $APP_CONST_FILE | %{$_ -replace "BUILT_AT_UNKNOWN_TIME", $DATE} | Out-File -Encoding "UTF8" $APP_CONST_FILE_TMP
mv -Force $APP_CONST_FILE_TMP $APP_CONST_FILE
