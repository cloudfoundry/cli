$ErrorActionPreference = "Stop"

. "$PSScriptRoot\integration-windows-setup.ps1"

echo "CF_INT_CLIENT_CREDENTIALS_TEST_MODE: $CF_INT_CLIENT_CREDENTIALS_TEST_MODE"

cd $Env:GOPATH\src\code.cloudfoundry.org\cli
ginkgo.exe -r -nodes=16 -flakeAttempts=2 -slowSpecThreshold=60 -randomizeAllSpecs ./integration/shared/isolated ./integration/v6/isolated ./integration/shared/plugin ./integration/v6/push
if ($LASTEXITCODE -gt 0)
{
	exit 1
}
ginkgo.exe -r -flakeAttempts=2 -slowSpecThreshold=60 -randomizeAllSpecs ./integration/shared/global ./integration/v6/global
if ($LASTEXITCODE -gt 0)
{
	exit 1
}
