$ErrorActionPreference = "Stop"

. "$PSScriptRoot\integration-windows-setup.ps1"

cd $Env:GOPATH\src\code.cloudfoundry.org\cli
ginkgo.exe -r -nodes=16 -flakeAttempts=2 -slowSpecThreshold=60 -randomizeAllSpecs ./integration/shared/experimental ./integration/v6/experimental
if ($LASTEXITCODE -gt 0)
{
	exit 1
}
