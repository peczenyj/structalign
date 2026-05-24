# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability within this project, please send an e-mail to **tiago.peczenyj+github@gmail.com**.

All security vulnerabilities will be promptly addressed. We request that you do not report security-related issues through public GitHub issues.

## Scope

`structalign` is a command-line developer tool that statically analyzes Go source
you point it at. It makes no network calls and only reads the files and packages
given as arguments — it never modifies them. The most relevant concerns are
therefore issues in how it parses or type-checks untrusted source (e.g. a crash on
malformed input). Reports of such issues are welcome via the e-mail above.
