$ErrorActionPreference = "Stop"
# in the future, this variable should cause PS to exit on non-zero exit codes from commands/exes (as opposed to PS cmdlets)
$PSNativeCommandUseErrorActionPreference = $true
# see https://github.com/PowerShell/PowerShell/issues/3415 and https://github.com/PowerShell/PowerShell-RFC/pull/277

# retrieved from https://docs.chocolatey.org/en-us/information/security
$chocoThumbprint = '83AC7D88C66CB8680BCE802E0F0F5C179722764B'

$scriptPath = (Get-Location).Path + '\installChocolatey.ps1'
(New-Object System.Net.WebClient).DownloadFile('https://chocolatey.org/install.ps1', $scriptPath)
(Get-AuthenticodeSignature .\installChocolatey.ps1).SignerCertificate.Thumbprint -eq $chocoThumbprint

Set-ExecutionPolicy Bypass -Scope Process
.\installChocolatey.ps1

choco install --no-progress -r -y innosetup --force

Get-Command iscc -ErrorAction Continue
