# expected environment variables
# SIGNING_KEY_WINDOWS_PFX
# SIGNING_KEY_WINDOWS_PASSPHRASE

param ($BinaryFilePath)

# add PATH to signtool.exe
$env:PATH="$env:PATH;C:\Program Files (x86)\Windows Kits\10\bin\x64"

[Convert]::FromBase64String($env:SIGNING_KEY_WINDOWS_PFX) | Set-Content -Path $env:RUNNER_TEMP\cert.pfx -AsByteStream

signtool sign /v /p "$env:SIGNING_KEY_WINDOWS_PASSPHRASE" /fd SHA256 /f "$env:RUNNER_TEMP\cert.pfx" "$BinaryFilePath"
