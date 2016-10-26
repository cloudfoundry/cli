SET GOPATH=%CD%\cf-release-repo
SET GATSPATH=%GOPATH%\src\github.com\cloudfoundry\cf-acceptance-tests

SET PATH=C:\Go\bin;%PATH%
SET PATH=C:\Program Files\Git\cmd\;%PATH%
SET PATH=%CD%\cf-release-repo\bin;%PATH%
SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=C:\Program Files\cURL\bin;%PATH%
SET PATH=%CD%;%PATH%

SET CONFIG=%CD%\cats-config\integration_config.json

go get -v github.com/onsi/ginkgo/ginkgo

pushd %CD%\cf-cli-binaries
	gzip -d cf-cli-binaries.tgz
	tar -xvf cf-cli-binaries.tar
	MOVE %CD%\cf-cli_winx64.exe ..\cf.exe
	dir ..
popd

cd %GATSPATH%
ginkgo.exe -slowSpecThreshold=120 -skip="NO_DEA_SUPPORT|go makes the app reachable via its bound route|SSO|takes effect after a restart, not requiring a push|doesn't die when printing 32MB|exercises basic loggregator|firehose data|Users can manage droplet bits for an app" -nodes=%NODES%
