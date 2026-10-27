# OpenDesk Free antivirus incident — 2026-10-27

**Status: unresolved, no false-positive determination.**

A Chrome download panel on Windows flagged **OpenDeskFree_Portable_0.6.3_preview1.exe** as *Đã phát hiện thấy vi-rút*. This is a virus warning, not simply a SmartScreen unfamiliar publisher notice.

## Verified provenance from public GitHub Releases
- v0.6.3 Preview 1 Portable SHA-256: `320e85304ea806234d89e892c5bf7451d57c5406ab7284933b5bff40e7f5e7a0`, size 13,038,080 bytes.
- v0.6.5 Preview 1 Portable SHA-256: `e9e0aa39fe48a879ae5c95a692ecc7997c8856b0767f44014fd0257bfdd2aa88`.
- v0.6.5 Preview 1 Setup SHA-256: `7096ef9a5243a28bf94c759b5d21591aeb0553b756aacd0f1e0383735fad6b9b`.
- All these releases were **unsigned previews**. A matching hash proves provenance/integrity against the release file, not that the code is free from malware.

## Source review finding
The Windows installer executes PowerShell `-ExecutionPolicy Bypass -EncodedCommand` to create a Start Menu shortcut and schedule self-cleanup. This behavior warrants redesign to reduce unnecessary security-sensitive actions, but the flagged asset is the **Portable** binary and does *not* bundle the Setup code. Therefore PowerShell in Setup is **not** a proven cause of this Chrome Portable warning.

## Release controls
- `scripts/defender-release-scan.ps1` runs a local Defender CustomScan of **both exact binaries** after Windows build and installer acceptance. No EXE execution during this scan.
- Scanner fails if antivirus is off, stale, absent, raises any detection, removes a binary, or cannot be checked. A JSON log is retained.
- GitHub release publisher still requires all three workflows successful on the **exact main HEAD**, with SHA-256 verification; an AV FAIL blocks publication.
- Antivirus scans are not equivalent to Chrome Safe Browsing checks or multi-vendor security review.

## Correct user / developer response
- Do **not** click Keep anyway, disable browser security, whitelist this app, or run the suspicious download.
- Inspect the exact detection in Chrome/Windows Security and provide name, detection product, and SHA-256 when available.
- Submit the **exact artifact** to the Microsoft developer malware analysis portal and, as relevant, Google's Safe Browsing feedback.
- Wait for an actual analyst/reputation verdict and evidence before claiming a false positive. Code-signing with a consistent publisher certificate helps trust but is not a guarantee against malware alerts.

Related issue: https://github.com/black11023/opendesk-free/issues/14

### CI environment observation

PR #15 first Windows CI run showed `AMRunningMode=Normal`, `AntivirusEnabled=True`, but `RealTimeProtectionEnabled=False` on a GitHub-hosted Windows runner. The strict initial gate correctly produced `BLOCKED/INCONCLUSIVE`. Updated diagnostic allows an on-demand scanner **only if** it detects the harmless EICAR positive control before scanning both EXEs; no protection exceptions or bypass are applied. If EICAR cannot be detected, release remains blocked.

### Independent Windows 11 ARM64 security runner

The `windows-latest` hosted runner on the initial PR showed realtime=False and did not detect EICAR in a second on-demand test. We therefore moved antivirus verification to a **separate Windows 11 ARM64 runner** (`windows-11-vs2026-arm`) after the x64 build completes. The runner downloads the exact artifacts from the build, verifies their SHA-256, requires EICAR detection, scans both x64 EXE files and fails the entire Windows workflow on any inconclusive security result. The source image documentation states Microsoft Defender is not disabled on the Windows 11 ARM images. Actual CI evidence, not this documentation alone, determines PASS/FAIL.
