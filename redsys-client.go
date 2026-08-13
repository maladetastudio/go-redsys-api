package redsys

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"net/url"
)

// Redsys Init this struct with your key to operate with the corresponding functions
type Redsys struct {
	Key string
}

// CreateMerchantParameters Return a string corresponding to a marshalled MerchantParametersRequest
func (r *Redsys) CreateMerchantParameters(data *MerchantParametersRequest) string {
	merchantMarshalledParams, _ := json.Marshal(data)
	return base64.URLEncoding.EncodeToString(merchantMarshalledParams)
}

// DecodeMerchantParameters Decode a response into a MerchantParametersResponse
func (r *Redsys) DecodeMerchantParameters(data string) MerchantParametersResponse {
	merchantParameters := MerchantParametersResponse{}
	decodedB64, _ := base64.URLEncoding.DecodeString(data)
	json.Unmarshal(decodedB64, &merchantParameters)
	unescapeInPlace(&merchantParameters)
	return merchantParameters
}

// unescapeInPlace reverses Redsys's per-field percent-encoding (seen on fields
// like Ds_Date "09%2F11%2F2015" and Ds_Hour "18%3A03"). It must run per-field,
// after JSON parsing, not on the whole blob beforehand: some fields legitimately
// contain a literal "%" that is not a valid escape sequence (e.g. DCC's
// Ds_Markup_DCC "0.0%"), and unescaping the whole JSON string at once lets one
// such field corrupt every other field silently.
func unescapeInPlace(m *MerchantParametersResponse) {
	fields := []*string{
		&m.Date, &m.Hour, &m.SecurePayment, &m.CardCountry, &m.Amount,
		&m.Currency, &m.Order, &m.MerchantCode, &m.Terminal, &m.Response,
		&m.MerchantData, &m.TransactionType, &m.ConsumerLanguage, &m.AuthorisationCode,
	}
	for _, f := range fields {
		if unescaped, err := url.QueryUnescape(*f); err == nil {
			*f = unescaped
		}
	}
}

// CreateMerchantSignature generates a merchant signature from MerchantParametersRequest
func (r *Redsys) CreateMerchantSignature(data *MerchantParametersRequest) string {
	stringMerchantParameters := r.CreateMerchantParameters(data)

	orderID := data.MerchantOrder

	encrypted := r.encrypt3DES(orderID)
	return r.mac256(stringMerchantParameters, encrypted)
}

// CreateMerchantSignatureNotif generates a signature for MerchantParametersResponse representing string
func (r *Redsys) CreateMerchantSignatureNotif(data string) string {
	merchantParametersResponse := r.DecodeMerchantParameters(data)

	orderID := merchantParametersResponse.Order
	encrypted := r.encrypt3DES(orderID)
	mac := r.mac256(data, encrypted)

	decodedMac, _ := base64.StdEncoding.DecodeString(mac)
	return base64.URLEncoding.EncodeToString(decodedMac)
}

// MerchantSignatureIsValid checks that two hmacs are equal
func (r *Redsys) MerchantSignatureIsValid(mac1 string, mac2 string) bool {
	return hmac.Equal([]byte(mac1), []byte(mac2))
}
