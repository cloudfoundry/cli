$ErrorActionPreference = "Stop"
trap { $host.SetShouldExit(1) }

echo "Work Directory: $pwd"
$Env:ROOT="$pwd"

$null = New-Item -ItemType Directory -Force -Path $Env:TEMP

# TODO: consider migrating choco to winget https://github.com/microsoft/winget-cli as preferred MS solution
if ((Get-Command "choco" -ErrorAction SilentlyContinue) -eq $null) {
  Set-ExecutionPolicy Bypass -Scope Process -Force
  [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
  $tempvar = (New-Object System.Net.WebClient).DownloadString('https://chocolatey.org/install.ps1')
  iex ($tempvar)
}

function Refresh-Choco-Env {
	Import-Module "C:\ProgramData\chocolatey\helpers\chocolateyProfile.psm1"
	refreshenv
	cd $Env:ROOT
}

Refresh-Choco-Env

$Env:GOPATH="$Env:ROOT\go"

$Env:PATH="$Env:HOME\go\bin;" + "$Env:PATH"
$Env:PATH="$Env:GOPATH\bin;" + "$Env:PATH"
$Env:PATH="$Env:GOROOT\bin;" + "$Env:PATH"
$Env:PATH="$pwd;" + "$Env:PATH"
$Env:PATH="$pwd\out;" + "$Env:PATH"

# This is for DEBUG
# function Get-Env-Info {
#   echo "Powershell: $((Get-Host).Version)"
#   echo "Working Directory: $pwd"
#   echo "GOPATH:            $Env:GOPATH"
#   echo "PATH:"
#   $Env:PATH.split(";")

#   echo "-------------"

#   Get-ChildItem Env: | Format-Table -Wrap -AutoSize
# }

# Get-Env-Info

$Env:RUN_ID=(openssl rand -hex 16)
$Env:GOFLAGS = "-mod=mod"

if ((Get-Command "ginkgo" -ErrorAction SilentlyContinue) -eq $null) {
	go install -v github.com/onsi/ginkgo/ginkgo@v1.16.4
}

$CF_INT_NAME=(Get-Content $pwd\metadata.json -Raw| Out-String | ConvertFrom-Json).name.trim()
$Env:CF_INT_PASSWORD=(Get-Content $pwd\cf-password -Raw).trim()
$Env:CF_INT_OIDC_PASSWORD=(Get-Content $pwd\uaa-oidc-password -Raw).trim()
$Env:CF_INT_OIDC_USERNAME="admin-oidc"
$Env:CF_INT_API="https://api.$CF_INT_NAME.cf-app.com"
$Env:CF_DIAL_TIMEOUT=15
# Enable SSL vaildation once toolsmiths supports it
# $Env:SKIP_SSL_VALIDATION="false"

Import-Certificate -Filepath "$pwd\$CF_INT_NAME.router.ca" -CertStoreLocation "cert:\LocalMachine\root"

New-Item "go/src/code.cloudfoundry.org" -Type Directory
New-Item -ItemType SymbolicLink -Path "$pwd/go/src/code.cloudfoundry.org/cli" -Target "$pwd"

cd go/src/code.cloudfoundry.org/cli
go install github.com/akavel/rsrc@v0.10.2

Get-Command make
Get-Item Makefile

make out/cf-cli_winx64.exe
Move-Item -Path $pwd\out\cf-cli_winx64.exe  -Destination $pwd\cf.exe -Force

cf.exe api $Env:CF_INT_API --skip-ssl-validation
cf.exe auth admin $Env:CF_INT_PASSWORD
cf.exe enable-feature-flag route_sharing

ginkgo.exe -r `
	-nodes=16 `
	-flakeAttempts=2 `
	-slowSpecThreshold=60 `
	-randomizeAllSpecs `
	./integration/shared/isolated `
	./integration/v7/isolated `
	./integration/shared/experimental `
	./integration/v7/experimental `
	./integration/v7/push

if ($LASTEXITCODE -gt 0)
{
	exit 1
}

ginkgo.exe -r `
	-flakeAttempts=2 `
	-slowSpecThreshold=60 `
	-randomizeAllSpecs `
	./integration/shared/global `
	./integration/v7/global

if ($LASTEXITCODE -gt 0)
{
	exit 1
}
