param ($InnoSetupConfig, $CfBinary, $InstallerOutput)

$ErrorActionPreference = "Stop"
# in the future, this variable should cause PS to exit on non-zero exit codes from commands/exes (as opposed to PS cmdlets)
$PSNativeCommandUseErrorActionPreference = $true
# see https://github.com/PowerShell/PowerShell/issues/3415 and https://github.com/PowerShell/PowerShell-RFC/pull/277

$innoSetupWorkDir = "$PSScriptRoot"
$licenseDir = "${PSScriptRoot}\..\license"

Move-Item -Force "$CfBinary" $innoSetupWorkDir\cf7.exe

# convert line-endings
Get-Content ${licenseDir}\LICENSE-WITH-3RD-PARTY-LICENSES | Set-Content "${innoSetupWorkDir}\LICENSE"
Get-Content ${licenseDir}\CF_NOTICE | Set-Content "${innoSetupWorkDir}\NOTICE"

iscc "$InnoSetupConfig"
Move-Item "${innoSetupWorkDir}\Output\mysetup.exe" "$InstallerOutput"

Get-ChildItem "${innoSetupWorkDir}\Output"
