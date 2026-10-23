# OpenDesk Free security

Security properties in v0.1.0:

- Application binds only to 127.0.0.1 with a random port.
- Host-header validation defends against common DNS rebinding attacks.
- Operations require a per-session random token and Same-Origin check.
- The app never sends uploaded files to an internet server.
- Uploaded files are processed only in memory, with total size and file count limits.
- Original documents are never overwritten.

Not a hardened sandbox: users should not process untrusted malformed ZIP/DOCX content on computers containing highly sensitive information until further fuzzing/security review is completed. Other programs running under the same Windows account can potentially access the locally served UI. The binaries are currently **unsigned**; GitHub source review and SHA256 checking are recommended.

Please avoid publicly sharing potential security exploits in Issues. Contact the maintainer using a private channel before disclosure.
