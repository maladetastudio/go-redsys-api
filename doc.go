// Package redsys builds, signs, decodes, and verifies the
// Ds_MerchantParameters/Ds_Signature pair used by Redsys's virtual payment
// gateway (TPV Virtual, the Spanish/Iberian card processor used by
// pagosonline.redsys.es). This package does no networking of its own — it
// only prepares the values a merchant sends and interprets the values Redsys
// sends back; the caller does the actual HTTP request or HTML form render.
//
// # Signing
//
// Two signature versions exist. New integrations should use
// [Redsys.CreateMerchantSignature512] (Ds_SignatureVersion
// [SignatureVersionSHA512], "HMAC_SHA512_V2": AES-128-CBC key diversification
// + HMAC-SHA512, verified against Redsys's own published worked example).
// [Redsys.CreateMerchantSignature] implements the legacy "HMAC_SHA256_V1"
// scheme (3DES key diversification + HMAC-SHA256), kept for merchants not yet
// migrated. Ds_MerchantParameters itself — [Redsys.CreateMerchantParameters],
// a base64url-encoded JSON blob — is identical either way; only the signing
// step differs.
//
//	r := redsys.Redsys{Key: merchantTerminalKey} // base64 key from Redsys's Admin Portal
//	params := &redsys.MerchantParametersRequest{
//		MerchantMerchantCode:    "999008881",
//		MerchantTerminal:        "1",
//		MerchantTransactionType: "0", // authorization
//		MerchantAmount:          "145",
//		MerchantCurrency:        "978", // EUR
//		MerchantOrder:           "1234567890",
//		MerchantUrlOK:           "https://example.com/pay/ok",
//		MerchantUrlKO:           "https://example.com/pay/ko",
//	}
//	dsMerchantParameters := r.CreateMerchantParameters(params)
//	dsSignature, err := r.CreateMerchantSignature512(params)
//
// # Integration types
//
// Redsys offers several ways to collect a payment; which
// [MerchantParametersRequest] fields matter, and what happens with the
// signed result, depends on which one a merchant uses:
//
//   - Redirection: the baseline flow. Send Ds_MerchantParameters/Ds_Signature/
//     Ds_SignatureVersion as three hidden fields in an HTML form that POSTs
//     to Redsys's hosted payment page (the customer's browser does the
//     navigation) — the merchant never touches the card data. Bizum rides on
//     this exact flow via MerchantPayMethods.
//   - InSite: Redsys's own JS library (redsysV3.js) renders card fields in
//     iframes it serves, embedded directly on the merchant's page — the
//     merchant's JS/DOM/server still never receive the PAN/CVV. The widget
//     hands back an opaque operation ID (DS_MERCHANT_IDOPER, valid ~30
//     minutes), which the merchant submits server-side via MerchantIdOper,
//     over a REST request to trataPeticionREST, to actually authorize the
//     payment.
//   - PayGold: a REST request (Ds_Merchant_TransactionType:
//     [TransactionTypePayGold]) that generates a payment link instead of
//     authorizing anything immediately; the link can be emailed/texted to a
//     customer to pay later. See the MerchantCustomer*/MerchantP2F* fields,
//     and [MerchantParametersResponse.URLPago2Fases] on the response.
//   - Apple Pay / Google Pay "direct integration": a wallet payment token
//     obtained client-side via Apple's/Google's own SDKs, submitted via the
//     XPayData/XPayType/XPayOrigen fields instead of relying on Redsys's
//     hosted wallet button.
//
// Both Redirection and InSite keep the merchant out of PCI-DSS scope: in
// both, Redsys itself is the only party that ever handles raw card data,
// whether on its own hosted page or inside iframes it serves on the
// merchant's page. This package's job is the same either way — build a
// correctly signed payload — but which fields you populate, and what you do
// with the result, differs per flow; see the type/field doc comments in
// structs.go for specifics.
//
// REST requests (refunds, PayGold link generation, and InSite's finalizing
// authorization) all POST the same three JSON fields
// (Ds_MerchantParameters/Ds_Signature/Ds_SignatureVersion) to
// trataPeticionREST and get back one JSON field, Ds_MerchantParameters,
// which you decode with [Redsys.DecodeMerchantParameters] or
// [Redsys.TryDecodeMerchantParameters].
//
// # Decoding responses and notifications
//
// Whatever Redsys sends back — a redirect flow's return leg, a REST call's
// reply, or an asynchronous online notification POSTed to your webhook URL —
// carries its own Ds_MerchantParameters/Ds_Signature pair. Decode the former
// with [Redsys.TryDecodeMerchantParameters] (preferred; surfaces decode
// errors) or [Redsys.DecodeMerchantParameters] (silently zero-value on
// error), and verify the latter by recomputing a signature over the raw,
// still-encoded Ds_MerchantParameters string — never over the decoded
// struct — with [Redsys.CreateMerchantSignatureNotif512] (or the legacy
// [Redsys.CreateMerchantSignatureNotif]), matching whichever
// Ds_SignatureVersion the sender declared, then comparing with
// [Redsys.MerchantSignatureIsValid]:
//
//	params := c.PostForm("Ds_MerchantParameters")
//	switch c.PostForm("Ds_SignatureVersion") {
//	case redsys.SignatureVersionSHA512:
//		expected, err := r.CreateMerchantSignatureNotif512(params)
//		// ... check err, then:
//		if !r.MerchantSignatureIsValid(expected, c.PostForm("Ds_Signature")) {
//			// reject: signature does not match
//		}
//	default:
//		// legacy HMAC_SHA256_V1, or unrecognized - decide how strict to be
//	}
//	event, err := r.TryDecodeMerchantParameters(params) // only after verifying
//
// Verify the signature before trusting any field of the decoded response -
// an unverified notification is just an unauthenticated HTTP POST.
package redsys
