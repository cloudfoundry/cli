$APP_CONST_FILE = $(split-path $MyInvocation.MyCommand.Definition) + "\..\src\cf\app_constants.go"
$APP_CONST_FILE_TMP = $APP_CONST_FILE + ".tmp"

get-content $APP_CONST_FILE | %{$_ -replace "SHA", $(git rev-parse --short HEAD)} | Out-File -Encoding "UTF8" $APP_CONST_FILE_TMP
mv -Force $APP_CONST_FILE_TMP $APP_CONST_FILE
