git submodule update --init

SET CLIPATH=c:\jenkins\workspace\go-cli-tests-windows32Bit
SET GOPATH=c:\Users\Administrator\go
go get -d github.com/cloudfoundry/cf-acceptance-tests/

cd %GOPATH%\src/github.com\cloudfoundry\cf-acceptance-tests

SET CATSPATH=%GOPATH%\src\github.com\cloudfoundry\cf-acceptance-tests
DEL c:\Users\Administrator\go\src\github.com\cloudfoundry\cf-acceptance-tests\cf.exe
bitsadmin.exe /transfer "DownloadStableCLI" https://s3.amazonaws.com/go-cli/builds/cf-windows-386.exe c:\Users\Administrator\go\src\github.com\cloudfoundry\cf-acceptance-tests\cf.exe

SET PATH=%PATH%;%CATSPATH%;C:\Program Files\cURL\bin

cd %CATSPATH%
git checkout master
SET CONFIG=%CATSPATH%\config.json

SET LOCAL_GOPATH=%CATSPATH%\Godeps\_workspace
MKDIR %LOCAL_GOPATH%\bin

SET GOPATH=%LOCAL_GOPATH%;%GOPATH%
SET PATH=%LOCAL_GOPATH%\bin;%PATH%

go install -v github.com/onsi/ginkgo/ginkgo
ginkgo -r -slowSpecThreshold=120 -skipPackage=diego -skip="go makes the app reachable via its bound route|SSO|takes effect after a restart, not requiring a push|doesn't die when printing 32MB|exercises basic loggregator"
