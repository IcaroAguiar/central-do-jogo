# Security Policy

## Supported versions

This project is pre-release. Security fixes target the default branch (`main`).

## Reporting a vulnerability

Please report security issues privately. Do **not** open a public GitHub issue for vulnerabilities that could expose users, credentials, or operator secrets.

Email the maintainer listed in the repository profile, or use GitHub Security Advisories when enabled for this repository.

Include:

- a short description of the issue and impact
- steps to reproduce (proof of concept without exploiting third parties)
- affected commit/branch if known

You should receive an acknowledgement within a few business days.

## Secrets and operator data

- Never commit secrets, tokens, cookies, OAuth client secrets, VAPID keys, or production credentials.
- Use environment variables / operator secret stores. `.env.example` contains non-sensitive placeholders only.
- Logs must not include tokens, cookies, passwords, or unnecessary personal data.
