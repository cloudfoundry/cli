[Setup]
ChangesEnvironment=yes
AlwaysShowDirOnReadyPage=yes
AppName=Cloud Foundry CLI
AppVersion=VERSION
AppVerName=Cloud Foundry CLI version VERSION
AppPublisher=Cloud Foundry Foundation
PrivilegesRequired=none
DefaultDirName={pf}\Cloud Foundry
SetupIconFile=cf.ico
UninstallDisplayIcon={app}\cf.ico

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: IsAdminLoggedOn and NeedsAddPath(ExpandConstant('{app}'))
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: not IsAdminLoggedOn and NeedsAddPath(ExpandConstant('{app}'))

[Files]
Source: LICENSE; DestDir: "{app}"
Source: NOTICE; DestDir: "{app}"
Source: cf7.exe; DestDir: "{app}"
Source: cf.ico; DestDir: "{app}"

[Run]
Filename: "{cmd}"; Parameters: "/C mklink ""{app}\cf.exe"" ""{app}\cf7.exe"""

[UninstallDelete]
Type: files; Name: "{app}\cf.exe"
Type: dirifempty; Name: "{app}"

[Code]
#include "common.iss"
