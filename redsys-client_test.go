package redsys

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test3DESEncryptionAndDecryption(t *testing.T) {
	const KEY = "Mk9m98IfEblmPfrpsawt7BmxObt98Jev"
	const DS_MERCHANT_ORDER = "1"
	const ENCRYPTED_TEXT = "Lr6bLJYWKrk="

	redsys := Redsys{Key: KEY}

	assert.Equal(t, ENCRYPTED_TEXT, redsys.encrypt3DES(DS_MERCHANT_ORDER), "Encryption result should be qual to "+ENCRYPTED_TEXT)

	assert.Equal(t, DS_MERCHANT_ORDER, redsys.decrypt3DES(ENCRYPTED_TEXT), "Decryption result should be qual to "+DS_MERCHANT_ORDER)

}

func TestSHA256Algorithm(t *testing.T) {
	const PARAMS = "eyJEU19NRVJDSEFOVF9BTU9VTlQiOiIxNDUiLCJEU19NRVJDSEFOVF9PUkRFUiI6IjEiLCJEU19NRVJDSEFOVF9NRVJDSEFOVENPREUiOiI5OTkwMDg4ODEiLCJEU19NRVJDSEFOVF9DVVJSRU5DWSI6Ijk3OCIsIkRTX01FUkNIQU5UX1RSQU5TQUNUSU9OVFlQRSI6IjAiLCJEU19NRVJDSEFOVF9URVJNSU5BTCI6Ijg3MSIsIkRTX01FUkNIQU5UX01FUkNIQU5UVVJMIjoiIiwiRFNfTUVSQ0hBTlRfVVJMT0siOiIiLCJEU19NRVJDSEFOVF9VUkxLTyI6IiJ9"
	const SIGNATURE = "3TEI5WyvHf1D/whByt1ENgFH/HPIP9UFuB6LkCYgj+E="
	const ENCRYPTED_KEY = "Lr6bLJYWKrk="

	redsys := Redsys{}
	assert.Equal(t, SIGNATURE, redsys.mac256(PARAMS, ENCRYPTED_KEY), "SHA256 result should be qual to "+SIGNATURE)
}

func TestMechantEncodingAndDecoding(t *testing.T) {
	const PARAMS = "eyJEc19NZXJjaGFudF9NZXJjaGFudENvZGUiOiI5OTkwMDg4ODEiLCJEc19NZXJjaGFudF9UZXJtaW5hbCI6Ijg3MSIsIkRzX01lcmNoYW50X1RyYW5zYWN0aW9uVHlwZSI6IjAiLCJEc19NZXJjaGFudF9BbW91bnQiOiIxNDUiLCJEc19NZXJjaGFudF9DdXJyZW5jeSI6Ijk3OCIsIkRzX01lcmNoYW50X09yZGVyIjoiMSJ9"
	const DS_MERCHANT_PARAMETERS = "eyJEc19EYXRlIjoiMDklMkYxMSUyRjIwMTUiLCJEc19Ib3VyIjoiMTglM0EwMyIsIkRzX1NlY3VyZVBheW1lbnQiOiIwIiwiRHNfQ2FyZF9Db3VudHJ5IjoiNzI0IiwiRHNfQW1vdW50IjoiMTQ1IiwiRHNfQ3VycmVuY3kiOiI5NzgiLCJEc19PcmRlciI6IjAwNjkiLCJEc19NZXJjaGFudENvZGUiOiI5OTkwMDg4ODEiLCJEc19UZXJtaW5hbCI6Ijg3MSIsIkRzX1Jlc3BvbnNlIjoiMDAwMCIsIkRzX01lcmNoYW50RGF0YSI6IiIsIkRzX1RyYW5zYWN0aW9uVHlwZSI6IjAiLCJEc19Db25zdW1lckxhbmd1YWdlIjoiMSIsIkRzX0F1dGhvcmlzYXRpb25Db2RlIjoiMDgyMTUwIn0="

	merchantParamsRequest := &MerchantParametersRequest{
		MerchantAmount:          "145",
		MerchantOrder:           "1",
		MerchantMerchantCode:    "999008881",
		MerchantCurrency:        "978",
		MerchantTransactionType: "0",
		MerchantTerminal:        "871",
		MerchantMerchantUrl:     "",
		MerchantUrlOK:           "",
		MerchantUrlKO:           "",
	}

	merchantParams := MerchantParametersResponse{
		Date:              "09/11/2015",
		Hour:              "18:03",
		SecurePayment:     "0",
		CardCountry:       "724",
		Amount:            "145",
		Currency:          "978",
		Order:             "0069",
		MerchantCode:      "999008881",
		Terminal:          "871",
		Response:          "0000",
		MerchantData:      "",
		TransactionType:   "0",
		ConsumerLanguage:  "1",
		AuthorisationCode: "082150",
	}

	redsys := Redsys{}

	assert.Equal(t, PARAMS, redsys.CreateMerchantParameters(merchantParamsRequest), "Create Merchant Parameters "+PARAMS)
	assert.Equal(t, merchantParams, redsys.DecodeMerchantParameters(DS_MERCHANT_PARAMETERS), "Decode Merchant Parameters "+PARAMS)
}

func TestDecodeMerchantParameters_DCCLiteralPercentDoesNotCorruptOtherFields(t *testing.T) {
	// Captured from a real Redsys refund response for a foreign-currency
	// (DCC) card. Ds_Markup_DCC contains a literal "0.0%" - a bare "%" not
	// followed by two hex digits - which previously made url.QueryUnescape
	// fail on the whole payload and silently zero out every field, including
	// Ds_Response (masking a successful refund as a failure).
	const dccMerchantParameters = "eyJEc19BbW91bnQiOiIzNjAwIiwiRHNfQ3VycmVuY3kiOiI5NzgiLCJEc19PcmRlciI6IjI2MDYwMWJrazdwZyIsIkRzX01lcmNoYW50Q29kZSI6IjM2MzE0NzU1NCIsIkRzX1Rlcm1pbmFsIjoiMiIsIkRzX1Jlc3BvbnNlIjoiMDkwMCIsIkRzX0F1dGhvcmlzYXRpb25Db2RlIjoiMDIxMDc2IiwiRHNfVHJhbnNhY3Rpb25UeXBlIjoiMyIsIkRzX1NlY3VyZVBheW1lbnQiOiIyIiwiRHNfTGFuZ3VhZ2UiOiIxIiwiRHNfTWVyY2hhbnREYXRhIjoiIiwiRHNfQ2FyZF9Db3VudHJ5IjoiODI2IiwiRHNfQ2FyZF9UeXBvbG9neSI6IkNPTlNVTU8iLCJEc19DYXJkX0JyYW5kIjoiMiIsIkRzX1Byb2Nlc3NlZFBheU1ldGhvZCI6Ijc5IiwiRHNfQ3VycmVuY3lfRENDIjoiODI2IiwiRHNfQW1vdW50X0RDQyI6IjMyODIiLCJEc19DdXJyZW5jeU5hbWVfRENDIjoiUE9VTkQgU1RFUkxJTkciLCJEc19NYXJrdXBfRENDIjoiMC4wJSIsIkRzX0V4Y2hhbmdlUmF0ZV9EQ0MiOiIxLjA5Njk1MSJ9"

	redsys := Redsys{}
	result := redsys.DecodeMerchantParameters(dccMerchantParameters)

	assert.Equal(t, "0900", result.Response, "a successful refund must decode with its real response code")
	assert.Equal(t, "260601bkk7pg", result.Order)
	assert.Equal(t, "021076", result.AuthorisationCode)
	assert.Equal(t, "3", result.TransactionType)
	assert.Equal(t, "3600", result.Amount)
	assert.Equal(t, "978", result.Currency)
}

func TestMerchantSignature(t *testing.T) {
	const KEY = "Mk9m98IfEblmPfrpsawt7BmxObt98Jev"

	merchantParamsRequest := &MerchantParametersRequest{
		MerchantAmount:          "145",
		MerchantOrder:           "1",
		MerchantMerchantCode:    "999008881",
		MerchantCurrency:        "978",
		MerchantTransactionType: "0",
		MerchantTerminal:        "871",
		MerchantMerchantUrl:     "",
		MerchantUrlOK:           "",
		MerchantUrlKO:           "",
	}
	redsys := Redsys{Key: KEY}
	const SIGNATURE = "FyetupQY42l5OuaBpazgN//z9veH6txWsUiYIAKE4FY="
	assert.Equal(t, SIGNATURE, redsys.CreateMerchantSignature(merchantParamsRequest), "Create Merchant Signature "+SIGNATURE)

	const RESPONSE_DS_MERCHANT_PARAMETERS = "eyJEc19EYXRlIjoiMDklMkYxMSUyRjIwMTUiLCJEc19Ib3VyIjoiMTglM0EwMyIsIkRzX1NlY3VyZVBheW1lbnQiOiIwIiwiRHNfQ2FyZF9Db3VudHJ5IjoiNzI0IiwiRHNfQW1vdW50IjoiMTQ1IiwiRHNfQ3VycmVuY3kiOiI5NzgiLCJEc19PcmRlciI6IjAwNjkiLCJEc19NZXJjaGFudENvZGUiOiI5OTkwMDg4ODEiLCJEc19UZXJtaW5hbCI6Ijg3MSIsIkRzX1Jlc3BvbnNlIjoiMDAwMCIsIkRzX01lcmNoYW50RGF0YSI6IiIsIkRzX1RyYW5zYWN0aW9uVHlwZSI6IjAiLCJEc19Db25zdW1lckxhbmd1YWdlIjoiMSIsIkRzX0F1dGhvcmlzYXRpb25Db2RlIjoiMDgyMTUwIn0="
	const RESPONSE_DS_SIGNATURE = "6DVpRPAPoChZh2cgaWnLqlfFsKeXdRfAO_tz-UrxJcU="
	assert.Equal(t, RESPONSE_DS_SIGNATURE, redsys.CreateMerchantSignatureNotif(RESPONSE_DS_MERCHANT_PARAMETERS), "Create Merchant Signature Notification "+RESPONSE_DS_SIGNATURE)

	assert.Equal(t, bool(true), redsys.MerchantSignatureIsValid(RESPONSE_DS_SIGNATURE, RESPONSE_DS_SIGNATURE), "Create Merchant Signature Notification")
}
