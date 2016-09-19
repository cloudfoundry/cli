function NeedsAddPath(Param: string): boolean;
var
  OrigPath: string;
begin
  if IsAdminLoggedOn then
  begin
    if not RegQueryStringValue(HKEY_LOCAL_MACHINE,
      'SYSTEM\CurrentControlSet\Control\Session Manager\Environment',
      'Path', OrigPath)
    then begin
      Result := True;
      exit;
    end;
   end
  else
  begin
    if not RegQueryStringValue(HKEY_CURRENT_USER,
      'Environment',
      'Path', OrigPath)
    then begin
      Result := True;
      exit;
    end;
  end;
  // look for the path with leading and trailing semicolon
  // Pos() returns 0 if not found
  Result := Pos(';' + Param + ';', ';' + OrigPath + ';') = 0;
end;

var
  OptionPage: TInputOptionWizardPage;

procedure InitializeWizard();
begin
  OptionPage :=
    CreateInputOptionPage(
      wpWelcome,
      'Choose installation options', 'Who should this application be installed for?',
      'Please select whether you wish to make this software available for all users or just yourself.',
      True, False);

  OptionPage.Add('&Anyone who uses this computer (run as administrator to enable)');
  OptionPage.Add('&Only for me');

  if IsAdminLoggedOn then
  begin
    OptionPage.Values[0] := True;
  end
    else
  begin
    OptionPage.Values[1] := True;
    OptionPage.CheckListBox.ItemEnabled[0] := False;
  end;
end;

function NextButtonClick(CurPageID: Integer): Boolean;
begin
  if CurPageID = OptionPage.ID then
  begin
    if OptionPage.Values[1] then
    begin
      // override the default installation to program files ({pf})
      WizardForm.DirEdit.Text := ExpandConstant('{userappdata}\Cloud Foundry')
    end
      else
    begin
      WizardForm.DirEdit.Text := ExpandConstant('{pf}\Cloud Foundry');
    end;
  end;
  Result := True;
end;
