package redsys

// MerchantParametersResponse struct to read Redsys API responses
type MerchantParametersResponse struct {
	Date               string `json:"Ds_Date"`
	Hour               string `json:"Ds_Hour"`
	SecurePayment      string `json:"Ds_SecurePayment"`
	CardCountry        string `json:"Ds_Card_Country,omitempty"`
	Amount             string `json:"Ds_Amount"`
	Currency           string `json:"Ds_Currency"`
	Order              string `json:"Ds_Order"`
	MerchantCode       string `json:"Ds_MerchantCode"`
	Terminal           string `json:"Ds_Terminal"`
	Response           string `json:"Ds_Response"`
	MerchantData       string `json:"Ds_MerchantData"`
	TransactionType    string `json:"Ds_TransactionType"`
	AuthorisationCode  string `json:"Ds_AuthorisationCode,omitempty"`
	ProcessedPayMethod string `json:"Ds_ProcessedPayMethod,omitempty"`
	ErrorCode          string `json:"Ds_ErrorCode,omitempty"`

	// ConsumerLanguage is tagged Ds_Language, not the "Ds_ConsumerLanguage"
	// this field was previously (and incorrectly) tagged as - confirmed
	// against two independent real Redsys payloads (a captured production
	// refund response, and Redsys's own PayGold REST documentation example).
	// The old tag meant this field silently decoded empty for every real
	// response; the field name is kept for API compatibility.
	ConsumerLanguage string `json:"Ds_Language,omitempty"`

	// CardNumber, CardBrand, and ExpiryDate are populated on a completed
	// authorization's response/notification (masked card number, brand
	// code, and expiry respectively) - confirmed against Redsys's PayGold
	// REST documentation, which shows them on the online notification sent
	// once a customer pays a generated link.
	CardNumber string `json:"Ds_CardNumber,omitempty"`
	CardBrand  string `json:"Ds_Card_Brand,omitempty"`
	ExpiryDate string `json:"Ds_ExpiryDate,omitempty"`

	// URLPago2Fases is the PayGold payment link generated for the customer
	// (only present when MerchantTransactionType was TransactionTypePayGold).
	// The Ds_AuthorisationCode/Ds_Response on this same response describe the
	// link-generation call, not a completed payment - the link hasn't been
	// paid yet.
	URLPago2Fases string `json:"Ds_UrlPago2Fases,omitempty"`

	// MerchantIdentifier is the generated card reference, present when the
	// request asked for one (see MerchantIdentifier on the request struct).
	MerchantIdentifier string `json:"Ds_Merchant_Identifier,omitempty"`
}

// MerchantParametersRequest struct to construct Redsys API requests
type MerchantParametersRequest struct {
	// Optional fields are tagged with omitempty
	MerchantMerchantCode       string `json:"Ds_Merchant_MerchantCode"`
	MerchantTerminal           string `json:"Ds_Merchant_Terminal"`
	MerchantTransactionType    string `json:"Ds_Merchant_TransactionType"`
	MerchantAmount             string `json:"Ds_Merchant_Amount"`
	MerchantCurrency           string `json:"Ds_Merchant_Currency"`
	MerchantOrder              string `json:"Ds_Merchant_Order"`
	MerchantMerchantUrl        string `json:"Ds_Merchant_MerchantURL,omitempty"`
	MerchantProductDescription string `json:"Ds_Merchant_ProductDescription,omitempty"`
	MerchantTitular            string `json:"Ds_Merchant_Titular,omitempty"`
	MerchantUrlOK              string `json:"Ds_Merchant_UrlOK,omitempty"`
	MerchantUrlKO              string `json:"Ds_Merchant_UrlKO,omitempty"`
	MerchantMerchantName       string `json:"Ds_Merchant_MerchantName,omitempty"`
	MerchantConsumerLanguage   string `json:"Ds_Merchant_ConsumerLanguage,omitempty"`
	MerchantPayMethods         string `json:"Ds_Merchant_Paymethods,omitempty"`

	// XPayData, XPayType, and XPayOrigen are for the Apple Pay / Google Pay
	// "direct integration" tier (a wallet payment token obtained client-side
	// via Apple's/Google's own SDKs, submitted directly instead of relying on
	// Redsys's hosted wallet button). Unused by the standard Redirection flow.
	XPayData   string `json:"DS_XPAYDATA,omitempty"`
	XPayType   string `json:"DS_XPAYTYPE,omitempty"`
	XPayOrigen string `json:"DS_XPAYORIGEN,omitempty"`

	// PayGold fields - only meaningful when MerchantTransactionType is
	// TransactionTypePayGold ("F"), sent via REST (trataPeticionREST), not
	// the Redirection flow. Redsys itself sends the SMS/email containing the
	// link when the corresponding contact field is set; leave both empty to
	// only generate the link (MerchantParametersResponse.URLPago2Fases) and
	// distribute it yourself through some other channel (e.g. WhatsApp,
	// which Redsys does not send via this API - only via its Admin Portal).
	MerchantCustomerMobile  string `json:"Ds_Merchant_Customer_Mobile,omitempty"`
	MerchantCustomerMail    string `json:"Ds_Merchant_Customer_Mail,omitempty"`
	MerchantP2FExpiryDate   string `json:"Ds_Merchant_P2F_ExpiryDate,omitempty"`
	MerchantCustomerSMSText string `json:"Ds_Merchant_Customer_Sms_Text,omitempty"`
	MerchantP2FXMLData      string `json:"Ds_Merchant_P2F_XMLData,omitempty"`

	// MerchantIdentifier requests (set to "REQUIRED") generation of a card
	// reference for the card used to pay a PayGold link; the reference comes
	// back in the online notification's MerchantIdentifier field, which
	// requires notifications to be configured in the Admin Portal.
	MerchantIdentifier string `json:"Ds_Merchant_Identifier,omitempty"`

	// MerchantIdOper is the InSite integration's operation identifier
	// (DS_MERCHANT_IDOPER, per Redsys's inSite documentation). Card data for
	// an InSite payment is entered into iframes served by Redsys's own JS
	// library (redsysV3.js), embedded on the merchant's page - the
	// merchant's JS/DOM/server never receive the PAN/CVV, only this opaque
	// operation ID, valid for 30 minutes. Set this field (alongside the
	// usual Ds_Merchant_Amount/Order/Currency/MerchantCode/Terminal and a
	// standard authorization Ds_Merchant_TransactionType) when finalizing an
	// InSite payment via a REST request to trataPeticionREST, the same REST
	// mechanism used by PayGold/refunds elsewhere in this package. Sign with
	// CreateMerchantSignature512 as usual.
	//
	// The exact casing/JSON key Redsys expects for this field in the REST
	// body (as opposed to the client-side redsysV3.js call, where it's
	// documented as DS_MERCHANT_IDOPER) was not confirmed against an
	// official PDF/library reference at the time this field was added -
	// verify it against Redsys's own integration docs and a sandbox
	// transaction before relying on it in production.
	MerchantIdOper string `json:"DS_MERCHANT_IDOPER,omitempty"`
}
