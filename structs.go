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
	ConsumerLanguage   string `json:"Ds_ConsumerLanguage,omitempty"`
	AuthorisationCode  string `json:"Ds_AuthorisationCode,omitempty"`
	ProcessedPayMethod string `json:"Ds_ProcessedPayMethod,omitempty"`
	ErrorCode          string `json:"Ds_ErrorCode,omitempty"`
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
}
