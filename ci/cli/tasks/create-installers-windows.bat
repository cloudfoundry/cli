SET ROOT_DIR=%CD%
SET ESCAPED_ROOT_DIR=%ROOT_DIR:\=\\%
SET /p VERSION=<%ROOT_DIR%\cli-ci\ci\VERSION

SET PATH=C:\Program Files\GnuWin32\bin;%PATH%
SET PATH=C:\Program Files (x86)\Inno Setup 5;%PATH%

SET PATH=C:\Program Files (x86)\Windows Kits\10\bin\x64;%PATH%

sed -i -e "s/VERSION/%VERSION%/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x64.iss
sed -i -e "s/CF_LICENSE/%ESCAPED_ROOT_DIR%\\LICENSE/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x64.iss
sed -i -e "s/CF_NOTICE/%ESCAPED_ROOT_DIR%\\NOTICE/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x64.iss
sed -i -e "s/CF_SOURCE/%ESCAPED_ROOT_DIR%\\cf.exe/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x64.iss
sed -i -e "s/CF_ICON/%ESCAPED_ROOT_DIR%\\cf.ico/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x64.iss

TYPE %ROOT_DIR%\cli-ci\ci\license\LICENSE-WITH-3RD-PARTY-LICENSES | MORE /P > LICENSE
TYPE %ROOT_DIR%\cli-ci\ci\license\NOTICE | MORE /P > NOTICE
COPY %ROOT_DIR%\cli-ci\ci\installers\windows\cf.ico cf.ico

MOVE %ROOT_DIR%\extracted-binaries\cf-cli_winx64.exe cf.exe

ISCC %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x64.iss

MOVE %ROOT_DIR%\cli-ci\ci\installers\windows\Output\mysetup.exe cf_installer.exe

zip %ROOT_DIR%\winstallers\cf-cli-installer_winx64.zip cf_installer.exe

sed -i -e "s/VERSION/%VERSION%/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x86.iss
sed -i -e "s/CF_LICENSE/%ESCAPED_ROOT_DIR%\\LICENSE/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x86.iss
sed -i -e "s/CF_NOTICE/%ESCAPED_ROOT_DIR%\\NOTICE/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x86.iss
sed -i -e "s/CF_SOURCE/%ESCAPED_ROOT_DIR%\\cf.exe/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x86.iss
sed -i -e "s/CF_ICON/%ESCAPED_ROOT_DIR%\\cf.ico/" %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x86.iss

MOVE %ROOT_DIR%\extracted-binaries\cf-cli_win32.exe cf.exe

ISCC %ROOT_DIR%\cli-ci\ci\installers\windows\windows-installer-x86.iss

MOVE %ROOT_DIR%\cli-ci\ci\installers\windows\Output\mysetup.exe cf_installer.exe

zip %ROOT_DIR%\winstallers\cf-cli-installer_win32.zip cf_installer.exe
