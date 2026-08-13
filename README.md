# go-redsys-api

Go client for Redsys's virtual payment gateway (TPV Virtual, the
Spanish/Iberian card processor behind `pagosonline.redsys.es`): builds and
signs `Ds_MerchantParameters`/`Ds_Signature`, and decodes/verifies Redsys's
responses and asynchronous notifications.

This package does **no networking of its own** — it prepares the values a
merchant sends and interprets the values Redsys sends back; the caller does
the actual HTTP request or HTML form render. See
[godoc](https://pkg.go.dev/github.com/maladetastudio/go-redsys-api) (or
`go doc .`) for the full API reference — this README covers setup and one
worked example per integration type.

Originally based on [GerardSoleCa/go-redsys-api](https://github.com/GerardSoleCa/go-redsys-api)
(itself based on [santiperez/node-redsys-api](https://github.com/santiperez/node-redsys-api)),
now maintained at `github.com/maladetastudio/go-redsys-api`.

## Installation

```
go get github.com/maladetastudio/go-redsys-api
```

## Quick start

```go
import "github.com/maladetastudio/go-redsys-api"

r := redsys.Redsys{Key: merchantTerminalKey} // base64 key from Redsys's Admin Portal
// or: r, err := redsys.NewRedsys(merchantTerminalKey) // validates the key up front

params := &redsys.MerchantParametersRequest{
	MerchantMerchantCode:    "999008881",
	MerchantTerminal:        "1",
	MerchantTransactionType: "0", // authorization
	MerchantAmount:          "145",
	MerchantCurrency:        "978", // ISO 4217 numeric, 978 = EUR
	MerchantOrder:           "1234567890",
	MerchantUrlOK:           "https://example.com/pay/ok",
	MerchantUrlKO:           "https://example.com/pay/ko",
}

dsMerchantParameters := r.CreateMerchantParameters(params)
dsSignature, err := r.CreateMerchantSignature512(params) // Ds_SignatureVersion: redsys.SignatureVersionSHA512
```

`dsMerchantParameters`/`dsSignature`/`redsys.SignatureVersionSHA512` are the
three values every integration type sends Redsys, one way or another — as
three hidden form fields (Redirection), or as three JSON fields in a REST
request body (refunds, PayGold, InSite's finalizing call).

## Signing versions

| | Ds_SignatureVersion | Key diversification | HMAC | Use |
|---|---|---|---|---|
| Current | `HMAC_SHA512_V2` (`redsys.SignatureVersionSHA512`) | AES-128-CBC | SHA-512 | `CreateMerchantSignature512` / `CreateMerchantSignatureNotif512` — new integrations |
| Legacy | `HMAC_SHA256_V1` | 3DES | SHA-256 | `CreateMerchantSignature` / `CreateMerchantSignatureNotif` — merchants not yet migrated |

Both known-answer tested against Redsys's own published worked examples (see
`redsys-client_test.go`). `Ds_MerchantParameters` itself
(`CreateMerchantParameters`) is identical either way — only the signing step
differs, and both live on the same `Redsys` value, so which one you call is
the only thing that changes.

## Integration types

Which fields of `MerchantParametersRequest` matter — and what you do with the
signed result — depends on which of Redsys's integration types you're using.
**Both Redirection and InSite keep the merchant out of PCI-DSS scope**: in
both, Redsys itself is the only party that ever handles raw card data, on its
own hosted page or inside iframes it serves on the merchant's page.

### Redirection (implemented)

The baseline flow. POST `Ds_MerchantParameters`/`Ds_Signature`/
`Ds_SignatureVersion` as three hidden fields in an HTML form targeting
Redsys's hosted payment page — the customer's browser does the navigation,
and the merchant never touches the card data.

```go
params := &redsys.MerchantParametersRequest{
	MerchantMerchantCode:    merchantCode,
	MerchantTerminal:        terminal,
	MerchantTransactionType: "0",
	MerchantAmount:          amount,
	MerchantCurrency:        "978",
	MerchantOrder:           orderID,
	MerchantUrlOK:           successURL,
	MerchantUrlKO:           cancelURL,
}
dsMerchantParameters := r.CreateMerchantParameters(params)
dsSignature, err := r.CreateMerchantSignature512(params)
// render an HTML form with these 3 values, POSTing to Redsys's payment URL
```

Bizum rides on this exact flow — set `MerchantPayMethods: "z"` to route the
customer straight to Bizum instead of the card form.

### InSite (implemented)

Redsys's own JS library (`redsysV3.js`) renders card fields in iframes it
serves, embedded directly on the merchant's page — the merchant's JS/DOM/
server still never receive the PAN/CVV. The widget calls back with an opaque
operation ID (`DS_MERCHANT_IDOPER`, valid ~30 minutes), which the merchant
then submits **server-side**, over a REST request to `trataPeticionREST`, to
actually authorize the payment:

```go
params := &redsys.MerchantParametersRequest{
	MerchantMerchantCode:    merchantCode,
	MerchantTerminal:        terminal,
	MerchantTransactionType: "0", // standard authorization
	MerchantAmount:          amount,
	MerchantCurrency:        "978",
	MerchantOrder:           orderID, // same order the widget was mounted with
	MerchantIdOper:          operationID, // from the client-side widget's callback
}
dsMerchantParameters := r.CreateMerchantParameters(params)
dsSignature, err := r.CreateMerchantSignature512(params)
// POST {Ds_MerchantParameters, Ds_Signature, Ds_SignatureVersion} to
// trataPeticionREST yourself; decode the JSON reply's Ds_MerchantParameters
// field with r.TryDecodeMerchantParameters
```

> **Unverified:** the exact REST field name/casing Redsys expects for the
> operation ID (`MerchantIdOper`/`DS_MERCHANT_IDOPER`) was transcribed from
> Redsys's inSite documentation but not confirmed against an official
> PDF/library reference or a live sandbox transaction at the time this field
> was added. Verify before relying on it in production.

### PayGold (groundwork only — not yet wired end-to-end by any consumer)

A REST request that generates a payment link instead of authorizing anything
immediately — the link can be emailed/texted to a customer to pay later.

```go
params := &redsys.MerchantParametersRequest{
	MerchantMerchantCode:    merchantCode,
	MerchantTerminal:        terminal,
	MerchantTransactionType: redsys.TransactionTypePayGold, // "F"
	MerchantAmount:          amount,
	MerchantCurrency:        "978",
	MerchantOrder:           orderID,
	// optional - leave both empty to just generate the link yourself:
	MerchantCustomerMobile: "666555444", // Redsys sends the SMS itself
	MerchantCustomerMail:   "customer@example.com", // Redsys sends the email itself
}
dsMerchantParameters := r.CreateMerchantParameters(params)
dsSignature, err := r.CreateMerchantSignature512(params)
// POST to trataPeticionREST; the reply's decoded MerchantIdentifier... er,
// URLPago2Fases field is the payment link.
```

Set `MerchantIdentifier: "REQUIRED"` to also request a reusable card
reference, delivered later in the online notification sent once the link is
paid (needs notifications configured in the Admin Portal).

### Apple Pay / Google Pay "direct integration" (groundwork only)

A wallet payment token obtained client-side via Apple's/Google's own SDKs,
submitted via `XPayData`/`XPayType`/`XPayOrigen` instead of relying on
Redsys's hosted wallet button. No consumer wires this up end-to-end yet.

## Decoding responses and notifications

Whatever Redsys sends back — a redirect flow's return leg, a REST call's
reply, or an asynchronous online notification POSTed to your webhook URL —
carries its own `Ds_MerchantParameters`/`Ds_Signature` pair.

Decode with `TryDecodeMerchantParameters` (preferred — surfaces decode
errors) or `DecodeMerchantParameters` (silently returns a zero-value response
on error). **Verify the signature before trusting any field**, over the raw,
still-encoded string — never the decoded struct:

```go
rawParams := c.PostForm("Ds_MerchantParameters")

var valid bool
switch c.PostForm("Ds_SignatureVersion") {
case redsys.SignatureVersionSHA512:
	expected, err := r.CreateMerchantSignatureNotif512(rawParams)
	valid = err == nil && r.MerchantSignatureIsValid(expected, c.PostForm("Ds_Signature"))
default: // legacy HMAC_SHA256_V1, or unrecognized
	expected := r.CreateMerchantSignatureNotif(rawParams)
	valid = r.MerchantSignatureIsValid(expected, c.PostForm("Ds_Signature"))
}
if !valid {
	// reject — an unverified notification is just an unauthenticated HTTP POST
}

event, err := r.TryDecodeMerchantParameters(rawParams) // only after verifying
```

`Ds_Response` codes in the `"0000"`–`"0099"` range mean success; everything
else is a decline or bank error, not a bug in your integration.

## Testing

```
go test ./...
```

Several tests are known-answer tests against Redsys's own published worked
examples (SHA-512 signing) or real captured payloads (field-tag regressions
— see `TestDecodeMerchantParameters_DCCLiteralPercentDoesNotCorruptOtherFields`,
`TestDecodeMerchantParameters_PayGoldPaidNotification`); prefer adding one of
these over a synthetic fixture when fixing a decoding bug, so the fix is
anchored to something Redsys actually sent.

## Consumers

Used by [`gitlab.com/maladetastudio/lueira/lueira`](https://gitlab.com/maladetastudio/lueira/lueira)
(`gateway/redsys.go`, `service/redsysservice/`) for the Redirection and REST
(refund/PayGold/InSite) flows.

## License

[MIT](LICENSE)
