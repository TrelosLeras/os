[Setup]
AppName=WhiteoutProjectOS
DefaultDirName={autopf}\WhiteoutProjectOS
DefaultGroupName=WhiteoutProjectOS
AppVersion=1.0.0
AppPublisher=Whiteout Project Team
AppPublisherURL=https://github.com/whiteout-project
SetupLogging=yes
CloseApplications=yes
RestartApplications=no
SetupIconFile=icon.ico 
WizardSmallImageFile=logo.png
OutputDir=userdocs
OutputBaseFilename=WhiteoutProjectOS_Installer
Compression=lzma
SolidCompression=yes

[Files]
; CRITICAL FIX: DestName changes the internal name so tasklist never finds the installer by mistake!
Source: "WhiteoutProjectOS.exe"; DestDir: "{app}"; DestName: "WhiteoutCore.exe"; Flags: ignoreversion
Source: "icon.ico"; DestDir: "{app}"; Flags: ignoreversion

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked

[UninstallDelete]
; Tells the uninstaller to clean up our manually created shortcuts
Type: files; Name: "{autoprograms}\WhiteoutProjectOS\WhiteoutProjectOS.lnk"
Type: dirifempty; Name: "{autoprograms}\WhiteoutProjectOS"
Type: files; Name: "{autodesktop}\WhiteoutProjectOS.lnk"

[Run]
Filename: "{app}\WhiteoutCore.exe"; Parameters: "-launch"; Description: "Launch WhiteoutProjectOS"; Flags: nowait postinstall skipifsilent

[UninstallRun]
Filename: "{app}\WhiteoutCore.exe"; Parameters: "-uninstall"; RunOnceId: "RemoveWSL"

[Code]
const
  WM_VSCROLL = $0115;
  SB_BOTTOM = 7;
  PM_REMOVE = 1;

var
  CredentialsPage: TInputQueryWizardPage;
  LogBox: TNewMemo; 

// --- HELPER: Windows Message Pump ---
type
  TMsg = record
    hwnd: HWND;
    message: UINT;
    wParam: Longint;
    lParam: Longint;
    time: DWORD;
    pt: TPoint;
  end;

function PeekMessage(var lpMsg: TMsg; hWnd: HWND; wMsgFilterMin, wMsgFilterMax, wRemoveMsg: UINT): BOOL; external 'PeekMessageW@user32.dll stdcall';
function TranslateMessage(const lpMsg: TMsg): BOOL; external 'TranslateMessage@user32.dll stdcall';
function DispatchMessage(const lpMsg: TMsg): Longint; external 'DispatchMessageW@user32.dll stdcall';
function GetTickCount: DWORD; external 'GetTickCount@kernel32.dll stdcall';

procedure AppProcessMessage;
var
  Msg: TMsg;
begin
  while PeekMessage(Msg, 0, 0, 0, PM_REMOVE) do
  begin
    TranslateMessage(Msg);
    DispatchMessage(Msg);
  end;
end;

// --- NEW HELPER: Smart Delay (60 FPS UI Waiter) ---
// Waits for the specified time while constantly pumping UI messages so animations don't freeze
procedure Delay(Milliseconds: DWORD);
var
  StartTick: DWORD;
begin
  StartTick := GetTickCount;
  while (GetTickCount - StartTick < Milliseconds) do
  begin
    AppProcessMessage;
    Sleep(10); // Tiny micro-sleep to prevent 100% CPU core lockup
  end;
end;

// --- HELPER: Native File Copy ---
function CloneLogFile(FileName: String): String;
var
  TempFile: String;
begin
  TempFile := ExpandConstant('{tmp}\read.log');
  FileCopy(FileName, TempFile, False); 
  Result := TempFile;
end;

// --- HELPER: Silent Process Monitor ---
function IsProcessRunning(const FileName: String): Boolean;
var
  WMIService, Processes: Variant;
begin
  Result := False;
  try
    WMIService := CreateOleObject('WbemScripting.SWbemLocator');
    Processes := WMIService.ConnectServer('localhost', 'root\CIMV2').ExecQuery('SELECT * FROM Win32_Process WHERE Name="' + FileName + '"');
    Result := Processes.Count > 0;
  except
    // Failsafe
  end;
end;

// --- HELPER: Read Last Line ---
function GetLastLine(FileName: String): String;
var
  Lines: TArrayOfString;
  TempLog: String;
begin
  TempLog := CloneLogFile(FileName);
  if LoadStringsFromFile(TempLog, Lines) then
  begin
    if GetArrayLength(Lines) > 0 then Result := Lines[GetArrayLength(Lines)-1]
    else Result := 'Waiting for log data...';
  end
  else Result := 'Initializing...';
end;

// --- HELPER: Abort Detector ---
function CheckForAbort(FileName: String): Boolean;
var
  Lines: TArrayOfString;
  TempLog: String;
  I: Integer;
begin
  Result := False;
  TempLog := CloneLogFile(FileName);
  if LoadStringsFromFile(TempLog, Lines) then
  begin
    for I := GetArrayLength(Lines) - 1 downto 0 do
    begin
      if Pos('[ABORT]', Lines[I]) > 0 then
      begin
        Result := True;
        Exit;
      end;
    end;
  end;
end;

// --- HELPER: Smart Progress Deep Scanner ---
function CalculateSmartProgress(FallbackPos: Integer): Integer;
var
  Lines: TArrayOfString;
  TempLog: String;
  I: Integer;
  LineText: String;
begin
  Result := FallbackPos;
  TempLog := CloneLogFile('C:\Windows\Temp\wos_install.log'); 
  if LoadStringsFromFile(TempLog, Lines) then
  begin
    for I := GetArrayLength(Lines) - 1 downto 0 do
    begin
      LineText := Lines[I];
      if Pos('9/9', LineText) > 0 then begin Result := 98; Exit; end;
      if Pos('8/9', LineText) > 0 then begin Result := 94; Exit; end;
      if Pos('7/9', LineText) > 0 then begin Result := 90; Exit; end;
      if Pos('6/9', LineText) > 0 then begin Result := 85; Exit; end;
      if Pos('5/9', LineText) > 0 then begin Result := 81; Exit; end;
      if Pos('4/9', LineText) > 0 then begin Result := 77; Exit; end;
      if Pos('3/9', LineText) > 0 then begin Result := 72; Exit; end;
      if Pos('2/9', LineText) > 0 then begin Result := 68; Exit; end;
      if Pos('1/9', LineText) > 0 then begin Result := 64; Exit; end;
    end;
  end;
end;

// --- HELPER: Create Shortcuts ---
procedure CreateShortcuts;
var
  WshShell, Shortcut: Variant;
  StartMenuPath, DesktopPath, TargetPath, IconPath: String;
begin
  WshShell := CreateOleObject('WScript.Shell');
  TargetPath := ExpandConstant('{app}\WhiteoutCore.exe');
  IconPath := ExpandConstant('{app}\icon.ico');

  StartMenuPath := ExpandConstant('{autoprograms}\WhiteoutProjectOS');
  ForceDirectories(StartMenuPath); 
  Shortcut := WshShell.CreateShortcut(StartMenuPath + '\WhiteoutProjectOS.lnk');
  Shortcut.TargetPath := TargetPath;
  Shortcut.Arguments := '-launch';
  Shortcut.IconLocation := IconPath;
  Shortcut.Save;

  if WizardIsTaskSelected('desktopicon') then
  begin
    DesktopPath := ExpandConstant('{autodesktop}\WhiteoutProjectOS.lnk');
    Shortcut := WshShell.CreateShortcut(DesktopPath);
    Shortcut.TargetPath := TargetPath;
    Shortcut.Arguments := '-launch';
    Shortcut.IconLocation := IconPath;
    Shortcut.Save;
  end;
end;

// --- BUILD THE CUSTOM UI PAGE ---
procedure InitializeWizard;
begin
  CredentialsPage := CreateInputQueryPage(wpWelcome,
    'Server Credentials', 'Please enter your background server details.',
    'Configure Linux Subsystem credentials or click next to use the default.');
    
  CredentialsPage.Add('Username:', False);
  CredentialsPage.Add('Password: (Default: Password123)', True);
  CredentialsPage.Add('Confirm Password:', True);
  CredentialsPage.Add('Hostname:', False);

  CredentialsPage.Values[0] := 'wp-os-user';
  CredentialsPage.Values[1] := ''; 
  CredentialsPage.Values[2] := '';
  CredentialsPage.Values[3] := 'wp-os-server';

  LogBox := TNewMemo.Create(WizardForm);
  LogBox.Parent := WizardForm.InstallingPage;
  LogBox.SetBounds(ScaleX(20), ScaleY(120), WizardForm.InstallingPage.ClientWidth - ScaleX(40), ScaleY(150));
  LogBox.ReadOnly := True;
  LogBox.Color := clBtnFace;
  LogBox.ScrollBars := ssNone; 
  LogBox.BorderStyle := bsSingle;
  LogBox.Visible := False; 
end;

function GetUsername(Param: String): String;
begin
  if Trim(CredentialsPage.Values[0]) = '' then Result := 'wp-os-user'
  else Result := CredentialsPage.Values[0];
end;

function GetPassword(Param: String): String;
begin
  if CredentialsPage.Values[1] = '' then Result := 'Password123'
  else Result := CredentialsPage.Values[1];
end;

function GetHostname(Param: String): String;
begin
  if Trim(CredentialsPage.Values[3]) = '' then Result := 'wp-os-server'
  else Result := CredentialsPage.Values[3];
end;

procedure CreateConfigFile;
var
  IniPath: String;
begin
  IniPath := ExpandConstant('{app}\config.ini');
  SetIniString('System', 'Username', GetUsername(''), IniPath);
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  Result := True;
  if CurPageID = CredentialsPage.ID then
  begin
    if CredentialsPage.Values[1] <> CredentialsPage.Values[2] then
    begin
      MsgBox('Passwords do not match. Please ensure both password fields are identical.', mbError, MB_OK);
      Result := False;
    end;
  end;
end;

// --- THE REAL INSTALLATION EXECUTION PHASE ---
procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode, Tick: Integer;
  Parameters, LogFile: String;
  CurrentLogLine, LastLogLine: String;
begin
  if CurStep = ssPostInstall then
  begin
    LogFile := 'C:\Windows\Temp\wos_install.log';
    
    // --- 1. THE BRUTAL ZOMBIE WIPE ---
    // Kill processes
    Exec(ExpandConstant('{sys}\taskkill.exe'), '/F /IM WhiteoutCore.exe /T', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    
    // Attempt 1 & 2: Native Delete & CMD Delete
    DeleteFile(LogFile);
    DeleteFile(ExpandConstant('{tmp}\read.log'));
    Exec(ExpandConstant('{sys}\cmd.exe'), '/c del /F /Q "' + LogFile + '"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    
    // Attempt 3: The Nuclear Option. Truncates the file to exactly 0 bytes even if it is locked!
    Exec(ExpandConstant('{sys}\cmd.exe'), '/c type NUL > "' + LogFile + '"', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    // ---------------------------------

    WizardForm.CancelButton.Enabled := False;

    CreateConfigFile;
    
    LogBox.Visible := True;
    LogBox.Lines.Clear; // Guarantee a completely blank UI slate
    
    WizardForm.StatusLabel.Caption := 'System Installation...';
    
    WizardForm.ProgressGauge.Min := 0;
    WizardForm.ProgressGauge.Max := 100;
    WizardForm.ProgressGauge.Position := 0;
    
    WizardForm.Refresh; 

    LastLogLine := ''; 
    // Now that the file is guaranteed empty, write the starting line
    SaveStringToFile(LogFile, '[System] Installer initialized...' + #13#10, False);

    // --- STAGE 1 & 2 ---
    WizardForm.StatusLabel.Caption := 'Enabling Windows features...';
    Exec(ExpandConstant('{app}\WhiteoutCore.exe'), '-stage1', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    WizardForm.ProgressGauge.Position := 15;
    
    Exec(ExpandConstant('{app}\WhiteoutCore.exe'), '-stage2', '', SW_HIDE, ewWaitUntilTerminated, ResultCode);
    WizardForm.ProgressGauge.Position := 30;

    // --- STAGE 3 ---
    WizardForm.StatusLabel.Caption := 'Preparing Linux Subsystem...';
    
    Exec(ExpandConstant('{app}\WhiteoutCore.exe'), '-stage3', '', SW_HIDE, ewNoWait, ResultCode);
    Delay(1000); 
    
    Tick := 0;
    while IsProcessRunning('WhiteoutCore.exe') do
    begin
      CurrentLogLine := GetLastLine(LogFile);
      if CurrentLogLine <> LastLogLine then
      begin
        LogBox.Lines.Add(CurrentLogLine);
        LogBox.SelStart := Length(LogBox.Text);
        LastLogLine := CurrentLogLine;
      end;
      
      Tick := Tick + 1;
      if (Tick mod 2 = 0) and (WizardForm.ProgressGauge.Position < 59) and (Pos('Installing Ubuntu', LogBox.Text) > 0) then
      begin
        WizardForm.ProgressGauge.Position := WizardForm.ProgressGauge.Position + 1;
      end;

      Delay(1000); 
    end;

    // --- ABORT CHECK ---
    if CheckForAbort(LogFile) then
    begin
      WizardForm.StatusLabel.Caption := 'Installation Aborted.';
      WizardForm.ProgressGauge.Position := 100;
      
      LogBox.Lines.Add(GetLastLine(LogFile));
      LogBox.SelStart := Length(LogBox.Text);
      
      MsgBox('Installation was cancelled by the user. Your existing Ubuntu environment was not modified.', mbError, MB_OK);
      WizardForm.CancelButton.Enabled := True; 
      Exit; 
    end;

    // --- STAGE 4 ---
    WizardForm.StatusLabel.Caption := 'Provisioning WhiteoutProjectOS...';
    WizardForm.ProgressGauge.Position := 60;
    
    Parameters := '-stage4 "' + GetUsername('') + '" "' + GetPassword('') + '" "' + GetHostname('') + '"';
    Exec(ExpandConstant('{app}\WhiteoutCore.exe'), Parameters, '', SW_HIDE, ewNoWait, ResultCode);
    Delay(1000); 
    
    while IsProcessRunning('WhiteoutCore.exe') do
    begin
      CurrentLogLine := GetLastLine(LogFile);
      if CurrentLogLine <> LastLogLine then
      begin
        LogBox.Lines.Add(CurrentLogLine);
        LogBox.SelStart := Length(LogBox.Text);
        LastLogLine := CurrentLogLine;
      end;
      
      WizardForm.ProgressGauge.Position := CalculateSmartProgress(WizardForm.ProgressGauge.Position);
      
      Delay(1000); 
    end;
    
    // --- WRAP UP ---
    WizardForm.ProgressGauge.Position := 100;
    WizardForm.StatusLabel.Caption := 'Finalizing shortcuts...';
    
    CreateShortcuts; 
    
    WizardForm.StatusLabel.Caption := 'Installation Complete.';
  end;
end;