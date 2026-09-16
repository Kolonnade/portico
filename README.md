# Portico

The entrance column of [Kolonnade](https://github.com/Kolonnade): a passkey-first identity provider you run yourself, and the libraries that sign people in to your sites.

Go and PostgreSQL, built on [`zitadel/oidc`](https://github.com/zitadel/oidc) so the OAuth and OpenID Connect machinery comes from a maintained implementation rather than from us.

> **Status: placeholder.** There is no code here yet. The provider is being rebuilt on `zitadel/oidc` before it is published, and this repository is the address it will land at. Nothing is released, and the first tag will be a `v0.x` with the API still moving.

## What it will do

- **Passkeys, and nothing to remember.** Enrolment, sign-in and recovery, with the credential-flag and sign-count rules synced passkeys actually need.
- **One sign-in across your sites**, whether they share a domain or not. No third-party cookies.
- **Several signed-in accounts at once**, numbered from 0 the way people expect, with each site signed in as whichever account the person picked for it.
- **Sign-out that arrives**, through OpenID Connect Back-Channel Logout, and profile changes pushed to sites as Shared Signals events instead of waiting for the next token refresh.
- **Apps as well as websites.** OpenID Connect Native SSO, so a family of apps on one device share a sign-in without sharing a refresh token, plus native passkey ceremonies. Swift first, in [`portico-swift`](https://github.com/Kolonnade); Android later.
- **Extension points where your product differs.** The profile — display name, avatar, whatever your product means by it — is an interface with a default implementation, not something baked in.
- **Conformance as a release gate**, not an aspiration: the OpenID Foundation's suite runs before a release goes out.

## What it will not do

Portico answers *who is this person*. It never answers *what may they do* — no roles, no permissions, no admin flags in the provider or in its tokens. Authorization belongs to your application, which is the only thing that knows what its own data means.

## When there is something to use

```
go get github.com/Kolonnade/portico
```

Until then, watching this repository for **Releases** is the way to hear about the first one.

## Security

Report vulnerabilities privately — see the [security policy](https://github.com/Kolonnade/.github/blob/main/SECURITY.md). Never in a public issue.

## License

MIT, once the code lands.
