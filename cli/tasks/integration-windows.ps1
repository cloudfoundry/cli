$ErrorActionPreference = "Stop"
$Env:GOPATH="$pwd\go"
$Env:CF_DIAL_TIMEOUT=15

$Env:PATH="C:\Go\bin;" + "$Env:PATH"
$Env:PATH="$Env:GOPATH\bin;" + "$Env:PATH"
$Env:PATH="C:\Program Files\GnuWin32\bin;" + "$Env:PATH"
$Env:PATH="$pwd;" + "$Env:PATH"

$DOMAIN=(Get-Content $pwd\bosh-lock\name -Raw).trim()
$Env:CF_INT_PASSWORD=(Get-Content $pwd\cf-credentials\cf-password -Raw).trim()
$Env:CF_INT_OIDC_PASSWORD=(Get-Content $pwd\cf-credentials\uaa-oidc-password -Raw).trim()
$Env:CF_INT_OIDC_USERNAME="admin-oidc"
$Env:CF_INT_API="https://api.$DOMAIN"
$Env:SKIP_SSL_VALIDATION="false"

$CF_INT_NAME = $DOMAIN.split(".")[0]
Import-Certificate -Filepath "$pwd\cf-credentials\cert_dir\$CF_INT_NAME.lb.cert" -CertStoreLocation "cert:\LocalMachine\root"

Import-Certificate -Filepath "$pwd\cf-credentials\cert_dir\$CF_INT_NAME.router.ca" -CertStoreLocation "cert:\LocalMachine\root"

pushd $pwd\cf-cli-binaries
	7z e cf-cli-binaries.tgz -y
	7z x cf-cli-binaries.tar -y
	Move-Item -Path $pwd\cf-cli_winx64.exe  -Destination ..\cf.exe -Force
popd

go get -v -u github.com/onsi/ginkgo/ginkgo

$Env:RUN_ID=(openssl rand -hex 16)

cd $Env:GOPATH\src\code.cloudfoundry.org\cli
ginkgo.exe -r -nodes=16 -flakeAttempts=2 -slowSpecThreshold=60 -randomizeAllSpecs ./integration/shared/isolated ./integration/v6/isolated ./integration/shared/plugin ./integration/v6/push
ginkgo.exe -r -flakeAttempts=2 -slowSpecThreshold=60 -randomizeAllSpecs ./integration/shared/global ./integration/v6/global
