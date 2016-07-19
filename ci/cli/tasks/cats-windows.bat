SET GOPATH=%CD%\cf-release-repo
SET GATSPATH=%GOPATH%\src\github.com\cloudfoundry\cf-acceptance-tests

SET PATH=C:\Go\bin;%PATH%
SET PATH=C:\Program Files\Git\cmd\;%PATH%
SET PATH=%CD%\cf-release-repo\bin;%PATH%
SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=C:\Program Files\cURL\bin;%PATH%
SET PATH=%CD%;%PATH%

SET /p DOMAIN=<%CD%\bosh-lite-lock\name
call %CD%\cli\ci\cli\tasks\create-cats-config.bat
SET CONFIG=%CD%\config.json

go get -v github.com/onsi/ginkgo/ginkgo

pushd %CD%\cf-cli-binaries
	gzip -d cf-cli-binaries.tgz
	tar -xvf cf-cli-binaries.tar
	MOVE %CD%\cf-cli_winx64.exe ..\cf.exe
	dir ..
popd

go get -v github.com/onsi/ginkgo/ginkgo

cd %GATSPATH%
ginkgo.exe -r -slowSpecThreshold=120 -skipPackage="logging,services,v3,routing_api,routing,backend_compatibility,ssh" -skip="NO_DEA_SUPPORT|go makes the app reachable via its bound route|SSO|takes effect after a restart, not requiring a push|doesn't die when printing 32MB|exercises basic loggregator|firehose data|Downloads the droplet for the app" -nodes=2
