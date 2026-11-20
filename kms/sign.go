package kms

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	HeaderDate            = "Date"
	HeaderApiVersion      = "x-kms-apiversion"
	HeaderApiName         = "x-kms-apiname"
	HeaderSignatureMethod = "x-kms-signaturemethod"
	HeaderAccessKeyId     = "x-kms-acccesskeyid"
	AlgoRsaPkcs1Sha256    = "RSA_PKCS1_SHA_256"
	ApiVersion20160102    = "2016-01-02"
)

func ExtractSignParams(param interface{}, tokenId string, action string) (queryParams, header map[string]string, err error) {
	queryParams = map[string]string{}
	data, err := json.Marshal(param)
	if err != nil {
		return nil, nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, nil, err
	}
	for k, v := range m {
		if v != nil && fmt.Sprintf("%v", v) != "" {
			queryParams[k] = fmt.Sprintf("%v", v)
		}
	}

	header = map[string]string{}
	header[HeaderAccessKeyId] = tokenId
	header[HeaderSignatureMethod] = AlgoRsaPkcs1Sha256
	header[HeaderDate] = time.Now().UTC().Format(time.RFC3339)
	header[HeaderApiName] = action
	header[HeaderApiVersion] = ApiVersion20160102
	return
}

func Sign(queryParams, header map[string]string, privateKey string) (string, error) {
	args := make(map[string]string)
	for k, v := range queryParams {
		args[k] = v
	}
	for k, v := range header {
		args[k] = v
	}
	toSignString := buildStringToSign(args)
	signature, err := Sha256WithRsa(toSignString, privateKey)
	if err != nil {
		return "", err
	}
	return signature, nil
}

func buildStringToSign(args map[string]string) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var buf bytes.Buffer
	for _, key := range keys {
		value := args[key]
		prefix := PercentEncode(key) + "="
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(prefix)
		buf.WriteString(PercentEncode(value))
	}
	return buf.String()
}

func PercentEncode(s string) string {
	spec := map[string]string{"+": "%20", "*": "%2A", "%7E": "~"}
	s = url.QueryEscape(s)
	for key, value := range spec {
		s = strings.ReplaceAll(s, key, value)
	}
	return s
}

func Sha256WithRsa(source, secret string) (signature string, err error) {
	block, _ := pem.Decode([]byte(secret))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	var privateKey interface{}
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		keyAny, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return "", fmt.Errorf("parse private key: pkcs1=%v pkcs8=%v", err, err2)
		}
		var ok bool
		privateKey, ok = keyAny.(*rsa.PrivateKey)
		if !ok {
			return "", fmt.Errorf("not RSA private key")
		}
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	h := crypto.Hash.New(crypto.SHA256)
	h.Write([]byte(source))
	hashed := h.Sum(nil)
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %v", err)
	}
	return base64.StdEncoding.EncodeToString(signatureBytes), nil
}
