# AGENTS.md

Instructions for AI coding agents working in this repository.

## What this is

A Go client library for Redsys's virtual payment gateway (TPV Virtual, the
Spanish/Iberian card processor at `pagosonline.redsys.es`). It builds and
signs `Ds_MerchantParameters`/`Ds_Signature`, and decodes/verifies Redsys's
responses and asynchronous notifications. **It does no networking of its
own** — callers do the actual HTTP request or HTML form render; this package
only prepares the values sent and interprets the values received.

Read `doc.go`'s package comment (`go doc .`) before making changes — it's
the canonical description of the signing schemes and integration types.

## Layout

- `structs.go` — `MerchantParametersRequest` (fields a merchant sends) and
  `MerchantParametersResponse` (fields Redsys sends back). One flat struct
  each, covering every integration type — see each field's own doc comment
  for which flow(s) it applies to.
- `redsys-client.go` — the public API: `Redsys` (holds a merchant's terminal
  key), parameter encoding/decoding, and both signature versions.
- `cipher.go` — low-level crypto: legacy 3DES key diversification + HMAC-
  SHA256 (`HMAC_SHA256_V1`), and current AES-128-CBC key diversification +
  HMAC-SHA512 (`HMAC_SHA512_V2`).
- `doc.go` — package-level doc comment only; no declarations.
- `redsys-client_test.go` — unit tests, several of them known-answer tests
  against Redsys's own published examples or real captured payloads.

## Build, test, lint

```
go build ./...
go vet ./...
gofmt -l .        # must produce no output
go test ./...
```

Run all four before considering any change done. There is no separate lint
config beyond `go vet`/`gofmt`.

## Conventions

- **Field names mirror Redsys's exact wire format.** JSON tags are the
  contract, not the Go field names — e.g. `MerchantMerchantCode` (an odd Go
  name) exists because Redsys's key is `Ds_Merchant_MerchantCode`. Never
  "clean up" a JSON tag to look more consistent; if it looks wrong, it's
  probably matching something Redsys actually sends (see the `Ds_Language`
  history below).
- **New fields get a doc comment stating which integration type/operation
  they belong to** and, ideally, a citation of where that was confirmed
  (a specific Redsys doc URL, or "confirmed against a real captured
  payload"). Grep the existing comments in `structs.go` for the pattern.
- **Don't guess a JSON key's exact casing/format and ship it silently.** If
  you can't verify a field against Redsys's official reference or a real
  request/response, say so explicitly in the doc comment (see
  `MerchantIdOper`'s comment for the template) rather than presenting it as
  confirmed. Redsys's field naming is inconsistent enough
  (`Ds_Merchant_MerchantCode` vs `DS_XPAYDATA` vs `Ds_UrlPago2Fases`) that
  guessing a plausible-looking name is a real way to ship a silently broken
  field — this has happened before in this codebase (`Ds_ConsumerLanguage`
  was wrong for years; the real field is `Ds_Language`, fixed in commit
  "Fix mistagged Ds_Language field, add missing card response fields" —
  the old tag meant the field silently decoded empty for every real
  response, an easy kind of bug to miss without a real-payload test).
- **Prefer testing against something Redsys actually sent** over a synthetic
  fixture — a known-answer test from Redsys's published docs (see
  `TestMAC512_RedsysPublishedExample`), or a real captured payload (see
  `TestDecodeMerchantParameters_DCCLiteralPercentDoesNotCorruptOtherFields`,
  `TestDecodeMerchantParameters_PayGoldPaidNotification`). A field-tag typo
  or a decode-corner-case bug is exactly the kind of thing a
  hand-constructed fixture won't catch, because the person constructing it
  has the same wrong mental model that caused the bug.
- **The signature covers the base64 string exactly as transmitted, never
  the decoded struct.** `decodeBase64Either`'s comment explains why
  `DecodeMerchantParameters` tolerating both base64 alphabets must never
  leak into signing/verification — decode tolerance is a parsing
  convenience, not license to normalize what gets signed.
- **Don't add networking, HTTP client code, or a `net/http` dependency to
  this package.** REST calls (refunds, PayGold, InSite's finalizing
  authorization) are the caller's job — this package hands back
  `Ds_MerchantParameters`/`Ds_Signature`/`Ds_SignatureVersion` and expects
  the caller to POST them and hand the reply's `Ds_MerchantParameters`
  field back to `TryDecodeMerchantParameters`. Keeping this package
  network-free is deliberate, not an oversight to "fix."
- **No new external dependencies** beyond `testify` (test-only). This is
  meant to stay small and dependency-light.

## PCI-DSS — why this package's shape matters

Redirection and InSite (the two implemented flows) keep a merchant out of
PCI-DSS scope because Redsys itself is the only party that ever handles raw
card data — on its own hosted page (Redirection) or inside iframes it serves
on the merchant's page (InSite). This package never has a field for a PAN,
CVV, or expiry date, and it should stay that way: if a change would require
adding one, stop and reconsider the approach rather than adding it — that
would mean card data is flowing through the merchant's own systems, which is
the exact thing these integration types exist to avoid.

## Versioning and downstream consumers

This module is consumed by `github.com/maladetastudio/lueira/lueira`
(GitLab: `gitlab.com/maladetastudio/lueira/lueira`) via a Go pseudo-version
pin in that repo's `go.mod` (e.g.
`github.com/maladetastudio/go-redsys-api v0.0.0-<timestamp>-<commit>`) —
there are no tagged releases yet. That means:

- **A commit only "ships" once a downstream `go.mod` is bumped to it.**
  Landing a change here does nothing on its own.
- **If you're working on a branch here and `master` moves underneath you
  (someone else merges first), rebase before pushing further work** — a
  downstream bump needs one commit containing everything, not two
  divergent lines of history. This has bitten this exact scenario before:
  a field addition based on an older `master` tip had to be rebased twice
  onto commits that landed in between, and a rebase conflict once
  silently resolved to the *wrong* side (re-picking an older pin over a
  newer one that was actually a strict superset) — when resolving a
  `go.mod`/`go.sum` conflict between two pseudo-version pins, check with
  `git merge-base --is-ancestor <old> <new>` which one is actually newer
  before picking a side; don't assume "ours" or "theirs" is correct.
- When bumping the downstream pin yourself, use
  `go get github.com/maladetastudio/go-redsys-api@<commit-sha>` from
  within that repo, then re-run its own build/vet/tests — a rename or
  field addition here can require a small mechanical fix on the consumer
  side (e.g. `UrlPago2Fases` → `URLPago2Fases`'s Go-naming-convention
  rename required a matching fix in that repo's `gateway/redsys.go`).

## Git branches in this repo

Feature branches on this repo have moved underneath in-progress work
mid-session before (see above). Before pushing new work or opening a PR,
`git fetch origin` and check whether your branch's base has moved
(`git merge-base --is-ancestor origin/master HEAD`) — rebase if not.
