git submodule update --init

SET CLIPATH=c:\jenkins\workspace\go-cli-tests-windows32Bit
SET GOPATH=%CLIPATH%

SET GOPATH=c:\Users\Administrator\go
go get -d github.com/cloudfoundry/cf-acceptance-tests/

cd %GOPATH%\src/github.com\cloudfoundry\cf-acceptance-tests
git pull

SET CATSPATH=%GOPATH%\src\github.com\cloudfoundry\cf-acceptance-tests
bitsadmin.exe /transfer "DownloadStableCLI" https://s3.amazonaws.com/go-cli/builds/cf-windows-386.exe c:\Users\Administrator\go\src\github.com\cloudfoundry\cf-acceptance-tests\gcf.exe

SET PATH=%PATH%;%CATSPATH%;C:\Program Files\cURL\bin

call %environment.bat

cd %CATSPATH%
SET CONFIG=%CATSPATH%\config.json
%GOPATH%\bin\ginkgo -r -v -slowSpecThreshold=300 -skip="admin buildpack"
