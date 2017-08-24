SET GOPATH=%CD%\go
SET CF_DIAL_TIMEOUT=15

SET PATH=C:\Go\bin;%PATH%
SET PATH=%GOPATH%\bin;%PATH%
SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=%CD%;%PATH%

SET /p DOMAIN=<%CD%\bosh-lite-lock\name
SET /p CF_PASSWORD=<%CD%\cf-credentials\cf-password
SET CF_API=https://api.%DOMAIN%

pushd %CD%\cf-cli-binaries
	gzip -d cf-cli-binaries.tgz
	tar -xvf cf-cli-binaries.tar
	MOVE %CD%\cf-cli_winx64.exe ..\cf.exe
popd

go get -v -u github.com/onsi/ginkgo/ginkgo

cd %GOPATH%\src\code.cloudfoundry.org\cli
ginkgo.exe -r -nodes=16 -flakeAttempts=2 -slowSpecThreshold=60 -randomizeAllSpecs ./integration/experimental
