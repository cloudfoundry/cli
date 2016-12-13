SET GOPATH=%CD%\go
SET GATSPATH=%GOPATH%\src\code.cloudfoundry.org\cli

SET PATH=C:\Go\bin;%PATH%
SET PATH=C:\Program Files\Git\cmd\;%PATH%
SET PATH=%GOPATH%\bin;%PATH%
SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=C:\Program Files\cURL\bin;%PATH%
SET PATH=%CD%;%PATH%

SET /p DOMAIN=<%CD%\bosh-lite-lock\name
SET CF_API=https://api.%DOMAIN%
call %CD%\cli-ci\ci\cli\tasks\create-cats-config.bat
SET CONFIG=%CD%\config.json

pushd %CD%\cf-cli-binaries
	gzip -d cf-cli-binaries.tgz
	tar -xvf cf-cli-binaries.tar
	MOVE %CD%\cf-cli_winx64.exe ..\cf.exe
popd

go get -v -u github.com/onsi/ginkgo/ginkgo

cd %GATSPATH%
ginkgo.exe -r -nodes=4 -flakeAttempts=2 -slowSpecThreshold=30 -randomizeSuites ./integration/isolated
ginkgo.exe -r -flakeAttempts=2 -slowSpecThreshold=30 -randomizeSuites ./integration/global
ginkgo.exe -r -flakeAttempts=2 -slowSpecThreshold=30 -randomizeSuites ./integration/plugin
