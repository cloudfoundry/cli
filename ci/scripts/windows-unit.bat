set root=%CD%

set /p VERSION=<version\number

call gopath\src\github.com\cloudfoundry\cli\bin\win_test.bat || exit /b 1

cd %root%

move /y gopath\src\github.com\cloudfoundry\cli\%CF_EXE_NAME% cf-windows%BITS%-%VERSION%.exe
