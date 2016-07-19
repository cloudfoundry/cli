SET GOPATH=%CD%\cf-release-repo
SET GATSPATH=%GOPATH%\src\github.com\cloudfoundry\cf-acceptance-tests
SET GOPATH=%GOPATH%;%GATSPATH%\Godeps\_workspace

SET PATH=C:\Go\bin;%PATH%
SET PATH=C:\Program Files\Git\cmd\;%PATH%
SET PATH=%CD%\cf-release-repo\bin;%PATH%
SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=C:\Program Files\cURL\bin;%PATH%
SET PATH=%CD%;%PATH%

SET /p DOMAIN=<%CD%\bosh-lite-lock\name
call %CD%\cli\ci\cli\tasks\create-cats-config.bat
SET CONFIG=%CD%\config.json

pushd %CD%\cf-cli-binaries
	gzip -d cf-cli-binaries.tgz
	tar -xvf cf-cli-binaries.tar
	MOVE %CD%\cf-cli_winx64.exe ..\cf.exe
popd

go get -v github.com/onsi/ginkgo/ginkgo

cd %GATSPATH%
ginkgo.exe -r -slowSpecThreshold=120 -nodes=2 -skip="NO_DIEGO_SUPPORT|Downloading droplets" . apps backend_compatibility detect docker internet_dependent security_groups ssh routing
