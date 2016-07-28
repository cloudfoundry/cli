SET ROOT_DIR=%CD%
SET ESCAPED_ROOT_DIR=%ROOT_DIR:\=\\%

SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=C:\Program Files (x86)\Inno Setup 5;%PATH%

REM You must have added to your Inno Setup a Signtool: (http://revolution.screenstepslive.com/s/revolution/m/10695/l/95041-signing-installers-you-create-with-inno-setup)
REM Name: "signtool", Command: "signtool.exe $p"
REM This is to add "signtool.exe" to your path, so it does not need to be fully qualified in the configuration above
SET PATH=C:\Program Files (x86)\Windows Kits\10\bin\x64;%PATH%

sed -i -e "s/VERSION/6.19.0/" %ROOT_DIR%\cli\installers\windows\windows-installer-x64.iss
sed -i -e "s/CF_SOURCE/%ESCAPED_ROOT_DIR%\\cf.exe/" %ROOT_DIR%\cli\installers\windows\windows-installer-x64.iss
sed -i -e "s/SIGNTOOL_CERT_PASSWORD/%SIGNTOOL_CERT_PASSWORD%/" %ROOT_DIR%\cli\installers\windows\windows-installer-x64.iss
sed -i -e "s/SIGNTOOL_CERT_PATH/%SIGNTOOL_CERT_PATH%/" %ROOT_DIR%\cli\installers\windows\windows-installer-x64.iss

pushd %ROOT_DIR%\64bit-windows-binary
	MOVE cf-cli_6.19.0_winx64 ..\cf.exe
popd

ISCC %ROOT_DIR%\cli\installers\windows\windows-installer-x64.iss

MOVE %ROOT_DIR%\cli\installers\windows\Output\mysetup.exe cf_installer.exe

zip %ROOT_DIR%\winstallers\cf-cli-installer_winx64.zip cf_installer.exe
