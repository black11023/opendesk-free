# OpenDesk Free V0.6.6 Preview 1 — Windows security remediation

This preview responds to an open report that Chrome blocked the v0.6.3 Portable download as a virus.

### Changed
- Replace Portable's external PowerShell startup-error handling and `rundll32.exe` browser launch with Windows `MessageBoxW` and `ShellExecuteW`.
- Generate installed Start Menu shortcuts using native COM, not encoded PowerShell.
- Remove encoded script execution and execution-policy bypass from delayed uninstaller cleanup.
- Keep native debug symbols in Windows binaries; preserve all free local features and source documents.
- Windows installer upgrade/uninstall test and live Portable tests required.
- Require actual ClamAV malware scanner to detect the benign EICAR sample and scan the two exact SHA-256-verified release EXEs.

**Important:** Antivirus PASS only means no detections by that engine and signatures at scan time. **Chrome Safe Browsing may still block unsigned binaries.** No determination has been made on the original 0.6.3 report; do not disable browser protection or antivirus. No software signing certificate has been issued.
