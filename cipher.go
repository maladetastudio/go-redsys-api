package redsys

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
)

var iv = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

var ivAES = make([]byte, aes.BlockSize)

func (r *Redsys) encrypt3DES(str string) string {

	block := getCipher(r.Key)
	cbc := cipher.NewCBCEncrypter(block, iv)

	decrypted := []byte(str)
	decryptedPadded, _ := zeroPad(decrypted, block.BlockSize())
	cbc.CryptBlocks(decryptedPadded, decryptedPadded)

	return base64.StdEncoding.EncodeToString(decryptedPadded)
}

func (r *Redsys) decrypt3DES(str string) string {

	block := getCipher(r.Key)
	cbc := cipher.NewCBCDecrypter(block, iv)

	encrypted, _ := base64.StdEncoding.DecodeString(str)
	cbc.CryptBlocks(encrypted, encrypted)

	unpaddedResult, _ := zeroUnpad(encrypted, block.BlockSize())

	return string(unpaddedResult)
}

func (r *Redsys) mac256(data string, key string) string {
	decodedKey, _ := base64.StdEncoding.DecodeString(key)
	hmac := hmac.New(sha256.New, []byte(decodedKey))
	hmac.Write([]byte(strings.TrimSpace(data)))
	result := hmac.Sum(nil)
	return base64.StdEncoding.EncodeToString(result)
}

// diversifyKeyAES computes the operation-specific key for the HMAC_SHA512_V2
// signature scheme: AES-128-CBC (zero IV, PKCS7 padding) of the order number,
// using the first 16 raw bytes of key (right-zero-padded if shorter) — NOT
// base64-decoded, unlike the legacy 3DES scheme's key handling. The result is
// standard (not URL-safe) base64, matching Redsys's own published example.
func diversifyKeyAES(key, order string) (string, error) {
	block, err := aes.NewCipher(aesKeyFrom(key))
	if err != nil {
		return "", err
	}

	cbc := cipher.NewCBCEncrypter(block, ivAES)
	padded := pkcs7Pad([]byte(order), aes.BlockSize)
	encrypted := make([]byte, len(padded))
	cbc.CryptBlocks(encrypted, padded)

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// aesKeyFrom returns the first 16 raw bytes of key, right-zero-padded to 16
// bytes if key is shorter.
func aesKeyFrom(key string) []byte {
	b := []byte(key)
	if len(b) >= 16 {
		return b[:16]
	}

	padded := make([]byte, 16)
	copy(padded, b)
	return padded
}

// mac512 computes the HMAC_SHA512_V2 signature: HMAC-SHA512 of data, keyed by
// the ASCII bytes of diversifiedKeyB64 (the base64 *string* itself, including
// its "=" padding — never base64-decoded, unlike the legacy mac256's key
// handling), encoded as unpadded base64url per Redsys's spec.
func mac512(data string, diversifiedKeyB64 string) string {
	h := hmac.New(sha512.New, []byte(diversifiedKeyB64))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

func getCipher(key string) cipher.Block {
	secretKey, err := base64.StdEncoding.DecodeString(key)

	if err != nil {
		log.Panic("Error decoding key", err)
	}

	crypto, err := des.NewTripleDESCipher(secretKey)
	if err != nil {
		log.Panic("Error generating cipher", err)
	}
	return crypto
}

// validateKey reports whether key is usable as a Redsys terminal key for the
// legacy 3DES signing path, without panicking — used by NewRedsys so a
// misconfigured key fails at construction time instead of inside getCipher
// mid-request.
func validateKey(key string) error {
	secretKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("redsys: key is not valid base64: %w", err)
	}

	if _, err := des.NewTripleDESCipher(secretKey); err != nil {
		return fmt.Errorf("redsys: key is not a valid 3DES key: %w", err)
	}

	return nil
}

// zeroPad function to terminate blocksize in zeros
func zeroPad(data []byte, blocklen int) ([]byte, error) {
	padlen := (blocklen - (len(data) % blocklen)) % blocklen
	pad := bytes.Repeat([]byte{0x00}, padlen)

	return append(data, pad...), nil
}

// zeroUnpad function to remove trailing zeros
func zeroUnpad(data []byte, blocklen int) ([]byte, error) {
	lastIndex := len(data)
	for lastIndex >= 0 && lastIndex > len(data)-blocklen-1 {
		lastIndex--
		if data[lastIndex] != 0 {
			break
		}
	}
	return data[:lastIndex+1], nil
}

// pkcs7Pad pads data to a multiple of blocklen per PKCS#7 (RFC 5652 §6.3),
// always adding at least one byte of padding, even if len(data) is already
// a multiple of blocklen.
func pkcs7Pad(data []byte, blocklen int) []byte {
	padlen := blocklen - (len(data) % blocklen)
	pad := bytes.Repeat([]byte{byte(padlen)}, padlen)
	return append(data, pad...)
}
