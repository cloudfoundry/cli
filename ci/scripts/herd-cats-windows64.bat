git submodule update --init

SET CLIPATH=c:\jenkins\workspace\go-cli-tests-windows64Bit
SET GOPATH=%CLIPATH%
c:\Go\bin\go build -v -o cf-windows-amd64.exe main

SET GOPATH=c:\Users\Administrator\go
SET CATSPATH=%GOPATH%\src\github.com\cloudfoundry\cf-acceptance-tests
copy %CLIPATH%\cf-windows-amd64.exe %CATSPATH%\gcf.exe /Y

SET PATH=%PATH%;%CATSPATH%;C:\Program Files\cURL\bin

call %environment.bat

cd %CATSPATH%
SET CONFIG=%CATSPATH%\config.json
%GOPATH%\bin\ginkgo -r -v -slowSpecThreshold=300 -skip="admin buildpack"
