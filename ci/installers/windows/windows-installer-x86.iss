[Setup]
ChangesEnvironment=yes
AppName=Cloud Foundry CLI
AppVersion=VERSION
AppVerName=Cloud Foundry CLI version VERSION
AppPublisher=Cloud Foundry Foundation
PrivilegesRequired=none
DefaultDirName={pf}\CloudFoundry
SetupIconFile=cf.ico
UninstallDisplayIcon={app}\cf.ico

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: IsAdminLoggedOn and NeedsAddPath(ExpandConstant('{app}'))
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; Check: not IsAdminLoggedOn and NeedsAddPath(ExpandConstant('{app}'))

[Files]
Source: CF_LICENSE; DestDir: "{app}"
Source: CF_NOTICE; DestDir: "{app}"
Source: CF_SOURCE; DestDir: "{app}"
Source: CF_ICON; DestDir: "{app}"

[Code]
#include "common.iss"
