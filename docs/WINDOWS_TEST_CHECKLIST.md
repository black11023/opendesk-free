# Windows 10/11 acceptance test — required before stable release

- [ ] Check `OpenDeskFree_Setup_0.1.0.exe` SHA-256 matches SHA256SUMS.txt.
- [ ] Scan installer and portable binary with Windows Defender.
- [ ] Install as a non-admin Windows 10 x64 user; verify no UAC elevation needed.
- [ ] Start Menu > OpenDesk Free opens browser dashboard on 127.0.0.1 with no command window.
- [ ] Select CSV, convert to XLSX; verify leading zeros, Vietnamese accents and formula strings are preserved as text.
- [ ] Organize mixed files: ZIP folder categories, manifest hashes, duplicate-name collision handling.
- [ ] Extract DOCX document body to UTF-8 TXT (formatting is intentionally not preserved).
- [ ] Upload duplicate files and verify SHA256 report; confirm original files did not change.
- [ ] Combine two UTF-8 TXT files and compare content.
- [ ] Verify max file count and 40MB operation limit error messages.
- [ ] Try bad Origin/token and confirm 403.
- [ ] Quit the app from UI and reopen it from the Start Menu.
- [ ] Reinstall same version while app is closed and verify executable is updated successfully.
- [ ] Uninstall using Settings > Installed apps; outputs in Downloads must remain unchanged.
- [ ] Test on an offline Windows PC; confirm all v0.1 tasks still work.
- [ ] Confirm Windows Defender/SmartScreen result; document unsigned publisher warning honestly.
