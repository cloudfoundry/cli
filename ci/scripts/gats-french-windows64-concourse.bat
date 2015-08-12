set BASE_DIR=%CD%

set BASE_GOPATH=%CD%\gopath

set GOPATH=%BASE_GOPATH%
set PATH=%BASE_GOPATH%\bin;%PATH%

set API_ENDPOINT=https://api.%BOSH_LITE_IP%.xip.io
set APPS_DOMAIN=%BOSH_LITE_IP%.xip.io
set ADMIN_USER=admin
set ADMIN_PASSWORD=admin
set CF_USER=user
set CF_PASSWORD=userpassword
set CF_ORG=cli-cats-org
set CF_SPACE=cli-cats-space

set CONFIG=%BASE_DIR%\config.json

echo {> %CONFIG%
echo "api": "%API_ENDPOINT%",>> %CONFIG%
echo "apps_domain": "%APPS_DOMAIN%",>> %CONFIG%
echo "admin_user": "%ADMIN_USER%",>> %CONFIG%
echo "admin_password": "%ADMIN_PASSWORD%",>> %CONFIG%
echo "cf_user": "%CF_USER%",>> %CONFIG%
echo "cf_user_password": "%CF_USER_PASSWORD%",>> %CONFIG%
echo "cf_org": "%CF_ORG%",>> %CONFIG%
echo "cf_space": "%CF_SPACE%",>> %CONFIG%
echo "skip_ssl_validation": true,>> %CONFIG%
echo "persistent_app_host": "persistent-app-win64",>> %CONFIG%
echo "default_timeout": 120,>> %CONFIG%
echo "cf_push_timeout": 210,>> %CONFIG%
echo "long_curl_timeout": 210,>> %CONFIG%
echo "broker_start_timeout": 330>> %CONFIG%
echo }>> %CONFIG%

set GATSPATH=%CD%\gopath\src\github.com\cloudfoundry\GATS

mkdir %BASE_GOPATH%\bin
move .\windows64-binary\cf* %BASE_GOPATH%\bin\cf.exe || exit /b 1

set GATS_DEPS_GOPATH=%GATSPATH%\Godeps\_workspace

set GOPATH=%GATS_DEPS_GOPATH%;%GOPATH%
set PATH=%GATS_DEPS_GOPATH%\bin;%PATH%;%GATSPATH%;C:\Program Files\cURL\bin

cd %GATSPATH%

go install github.com/onsi/ginkgo/ginkgo || exit /b 1

ginkgo -r -slowSpecThreshold=120 ./translations
