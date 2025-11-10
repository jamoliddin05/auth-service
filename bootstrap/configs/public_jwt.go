package configs

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
)

type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func LoadJWKSFromPEM(pemStr string, kid string) (string, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return "", errors.New("failed to parse PEM block")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse RSA public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("not an RSA public key")
	}

	// Convert modulus and exponent to base64 URL encoding
	n := base64.RawURLEncoding.EncodeToString(rsaPub.N.Bytes())

	// Convert exponent (int) to bytes
	eBytes := make([]byte, 4)
	eLen := 0
	for exp := rsaPub.E; exp > 0; exp >>= 8 {
		eBytes[3-eLen] = byte(exp & 0xff)
		eLen++
	}
	e := base64.RawURLEncoding.EncodeToString(eBytes[4-eLen:])

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		Kid: kid,
		N:   n,
		E:   e,
	}

	jwks := JWKS{
		Keys: []JWK{jwk},
	}

	data, err := json.Marshal(jwks)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
