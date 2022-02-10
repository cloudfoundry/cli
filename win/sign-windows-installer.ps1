# expected environment variables
# SIGNING_KEY_WINDOWS_PFX
# SIGNING_KEY_WINDOWS_PASSPHRASE

# add PATH to signtool.exe
$env:PATH="$env:PATH;C:\Program Files (x86)\Windows Kits\10\bin\x64"
[Convert]::FromBase64String($env:SIGNING_KEY_WINDOWS_PFX) | Set-Content -Path $env:RUNNER_TEMP\cert.pfx -AsByteStream
pwd
Get-ChildItem win\Output

Get-Command signtool
# signtool sign /v /fd SHA256 /f $env:RUNNER_TEMP.pfx /p "$SIGNING_KEY_WINDOWS_PASSPHRASE" .\dist\hello-windows_windows_amd64\hello.exe
signtool sign /v /p "$env:SIGNING_KEY_WINDOWS_PASSPHRASE" /fd SHA256 /f $env:RUNNER_TEMP\cert.pfx $env:RUNNER_TEMP\x64\cf8_installer.exe
