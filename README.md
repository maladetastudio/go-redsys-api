# go-redsys-api

Go client for Redsys's virtual payment gateway (TPV Virtual): builds and signs
`Ds_MerchantParameters`/`Ds_Signature`, and decodes/verifies Redsys's
responses and asynchronous notifications.

Originally based on [GerardSoleCa/go-redsys-api](https://github.com/GerardSoleCa/go-redsys-api)
(itself based on [santiperez/node-redsys-api](https://github.com/santiperez/node-redsys-api)),
now maintained at `github.com/maladetastudio/go-redsys-api`.

## Installation

	go get github.com/maladetastudio/go-redsys-api

## Tests

	go test ./...

## Signing

`CreateMerchantSignature512`/`CreateMerchantSignatureNotif512` implement
Redsys's current standard (`Ds_SignatureVersion: HMAC_SHA512_V2` - AES-128-CBC
key diversification + HMAC-SHA512). `CreateMerchantSignature`/
`CreateMerchantSignatureNotif` implement the legacy `HMAC_SHA256_V1` scheme,
kept for merchants not yet migrated.

## Integration types

Redsys offers several ways to collect a payment; this package's request/
response structs (`structs.go`) support:

- **Redirection** (implemented): the standard flow - build
  `Ds_MerchantParameters`/`Ds_Signature` with `CreateMerchantParameters`/
  `CreateMerchantSignature512`, and have the customer's browser POST them to
  Redsys's hosted payment page. Bizum rides on this same flow via
  `Ds_Merchant_Paymethods`.
- **InSite** (implemented): card fields are rendered in iframes served by
  Redsys's own JS library (`redsysV3.js`), embedded directly on the
  merchant's page - card data never reaches the merchant's JS/DOM/server.
  The client-side widget returns an opaque operation ID
  (`DS_MERCHANT_IDOPER`), which the merchant then submits server-side, via a
  REST request to `trataPeticionREST`, using `MerchantIdOper` on
  `MerchantParametersRequest`. This (like Redirection) keeps the merchant out
  of PCI-DSS scope, since raw card data is exclusively handled by Redsys.
- **PayGold** (groundwork only, not yet wired end-to-end by any consumer):
  generates a payment link via REST (`Ds_Merchant_TransactionType`:
  `TransactionTypePayGold`) that can be sent to a customer out-of-band
  (SMS/email/etc). See the `MerchantCustomer*`/`MerchantP2F*` fields.
- **Apple Pay / Google Pay "direct integration"** (groundwork only): the
  `XPayData`/`XPayType`/`XPayOrigen` fields, for a wallet payment token
  obtained client-side via Apple's/Google's own SDKs.

## License

[MIT](LICENSE)
