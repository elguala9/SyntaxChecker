; SyntaxChecker - Inno Setup script
; Build: iscc installer.iss  (oppure: make installer)
; Produces: dist\SyntaxChecker-Setup.exe

#define MyAppName        "SyntaxChecker"
#define MyAppPublisher   "Parresia"
#define MyAppExeName     "syntaxchecker-mcp.exe"
#ifndef MyAppVersion
  #define MyAppVersion   "0.1.0"
#endif

[Setup]
AppId={{F2A8E2C3-3B1D-4B7F-9C8E-9B7B1F2A8E2C}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableProgramGroupPage=yes
OutputDir=dist
OutputBaseFilename=SyntaxChecker-Setup
Compression=lzma2
SolidCompression=yes
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
PrivilegesRequired=lowest
PrivilegesRequiredOverridesAllowed=dialog
WizardStyle=modern
UninstallDisplayName={#MyAppName} {#MyAppVersion}
ChangesEnvironment=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "italian"; MessagesFile: "compiler:Languages\Italian.isl"

[Tasks]
Name: "addtopath"; Description: "Aggiungi la cartella di installazione al PATH utente"; GroupDescription: "Integrazione sistema:"

[Files]
Source: "dist\syntax-checker.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "dist\syntaxchecker-mcp.exe";  DestDir: "{app}"; Flags: ignoreversion
Source: "for-ia.md";            DestDir: "{app}\docs"; Flags: ignoreversion
Source: "for-agent.md";         DestDir: "{app}\docs"; Flags: ignoreversion

[Icons]
Name: "{group}\Documentazione (for-ia)";    Filename: "{app}\docs\for-ia.md"
Name: "{group}\Documentazione (for-agent)"; Filename: "{app}\docs\for-agent.md"
Name: "{group}\Disinstalla {#MyAppName}";   Filename: "{uninstallexe}"

[Code]
const
  EnvironmentKeyUser   = 'Environment';
  EnvironmentKeySystem = 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';

function GetEnvHive(): Integer;
begin
  if IsAdminInstallMode then
    Result := HKEY_LOCAL_MACHINE
  else
    Result := HKEY_CURRENT_USER;
end;

function GetEnvSubkey(): string;
begin
  if IsAdminInstallMode then
    Result := EnvironmentKeySystem
  else
    Result := EnvironmentKeyUser;
end;

function PathContains(const PathList, NewItem: string): Boolean;
var
  Haystack, Needle: string;
begin
  Haystack := ';' + Lowercase(PathList) + ';';
  Needle   := ';' + Lowercase(NewItem) + ';';
  Result   := Pos(Needle, Haystack) > 0;
end;

procedure AddToPath(const Dir: string);
var
  CurrentPath: string;
begin
  if not RegQueryStringValue(GetEnvHive, GetEnvSubkey, 'Path', CurrentPath) then
    CurrentPath := '';
  if PathContains(CurrentPath, Dir) then
    Exit;
  if (CurrentPath <> '') and (CurrentPath[Length(CurrentPath)] <> ';') then
    CurrentPath := CurrentPath + ';';
  CurrentPath := CurrentPath + Dir;
  RegWriteExpandStringValue(GetEnvHive, GetEnvSubkey, 'Path', CurrentPath);
end;

procedure RemoveFromPath(const Dir: string);
var
  CurrentPath, Rebuilt, Part: string;
  P: Integer;
begin
  if not RegQueryStringValue(GetEnvHive, GetEnvSubkey, 'Path', CurrentPath) then
    Exit;
  Rebuilt := '';
  CurrentPath := CurrentPath + ';';
  repeat
    P := Pos(';', CurrentPath);
    if P = 0 then Break;
    Part := Copy(CurrentPath, 1, P - 1);
    CurrentPath := Copy(CurrentPath, P + 1, Length(CurrentPath));
    if (Part <> '') and (Lowercase(Part) <> Lowercase(Dir)) then begin
      if Rebuilt <> '' then Rebuilt := Rebuilt + ';';
      Rebuilt := Rebuilt + Part;
    end;
  until CurrentPath = '';
  RegWriteExpandStringValue(GetEnvHive, GetEnvSubkey, 'Path', Rebuilt);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then begin
    if WizardIsTaskSelected('addtopath') then
      AddToPath(ExpandConstant('{app}'));
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
    RemoveFromPath(ExpandConstant('{app}'));
end;
