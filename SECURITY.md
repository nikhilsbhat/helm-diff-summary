# Security Policy

## Reporting A Vulnerability

Please report security issues privately by emailing the maintainer listed on the GitHub profile for `nikhilsbhat`.

Avoid opening public issues for vulnerabilities involving secret handling, command execution, release artifacts, or notification webhooks.

## Secret Handling

Notification credentials are read from environment variables instead of command-line flags so they are less likely to appear in shell history, CI logs, or process lists.
