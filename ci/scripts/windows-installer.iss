[Setup]
AppName=Cloud Foundry CLI
AppVersion=VERSION
DefaultDirName={pf}\CloudFoundry

[Registry]
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"

[Files]
Source: CF_SOURCE; DestDir: "{app}"
