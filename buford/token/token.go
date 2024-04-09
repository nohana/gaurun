// Package token
// original: https://github.com/sideshow/apns2/blob/master/token/token.go
// Copyright (c) 2016 Adam Jones
package token

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"github.com/mercari/gaurun/gaurun"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	// TokenTimeout is the period of time in seconds that a token is valid for.
	// If the timestamp for token issue is not within the last hour, APNs
	// rejects subsequent push messages. This is set to under an hour so that
	// we generate a new token before the existing one expires.
	TokenTimeout = 3000
)

// Possible errors when parsing a .p8 file.
var (
	ErrAuthKeyNotPem   = errors.New("token: AuthKey must be a valid .p8 PEM file")
	ErrAuthKeyNotECDSA = errors.New("token: AuthKey must be of type ecdsa.PrivateKey")
	ErrAuthKeyNil      = errors.New("token: AuthKey was nil")
)

// Token represents an Apple Provider Authentication Token (JSON Web Token).
type Token struct {
	sync.Mutex
	AuthKey  *ecdsa.PrivateKey
	KeyID    string
	TeamID   string
	IssuedAt int64
	Bearer   string
}

func AuthKeyFromConfig(iosConfig gaurun.SectionIos) (*ecdsa.PrivateKey, error) {
	if len(iosConfig.TokenAuthKeyBase64) > 0 {
		return AuthKeyFromBase64(iosConfig.TokenAuthKeyBase64)
	}
	return AuthKeyFromFile(iosConfig.TokenAuthKeyPath)
}

// AuthKeyFromFile loads a .p8 certificate from a local file and returns a
func AuthKeyFromFile(filename string) (*ecdsa.PrivateKey, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return AuthKeyFromBytes(bytes)
}

func AuthKeyFromBase64(authKeyBase64 string) (*ecdsa.PrivateKey, error) {
	// Base64デコード
	authKey, err := base64.StdEncoding.DecodeString(authKeyBase64)
	if err != nil {
		return nil, err
	}
	return AuthKeyFromBytes([]byte(authKey))
}

// AuthKeyFromBytes loads a .p8 certificate from an in memory byte array and
func AuthKeyFromBytes(bytes []byte) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, ErrAuthKeyNotPem
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	switch pk := key.(type) {
	case *ecdsa.PrivateKey:
		return pk, nil
	default:
		return nil, ErrAuthKeyNotECDSA
	}
}

// GenerateBearerIfExpired checks to see if the token is about to expire and
// generates a new token.
func (t *Token) GenerateBearerIfExpired() (bearer string) {
	t.Lock()
	defer t.Unlock()
	if t.Expired() {
		// TODO: error handling
		t.Generate()
	}
	return t.Bearer
}

// Expired checks to see if the token has expired.
func (t *Token) Expired() bool {
	return time.Now().Unix() >= (t.IssuedAt + TokenTimeout)
}

// Generate creates a new token.
func (t *Token) Generate() (bool, error) {
	if t.AuthKey == nil {
		return false, ErrAuthKeyNil
	}
	issuedAt := time.Now().Unix()
	jwtToken := &jwt.Token{
		Header: map[string]interface{}{
			"alg": "ES256",
			"kid": t.KeyID,
		},
		Claims: jwt.MapClaims{
			"iss": t.TeamID,
			"iat": issuedAt,
		},
		Method: jwt.SigningMethodES256,
	}
	bearer, err := jwtToken.SignedString(t.AuthKey)
	if err != nil {
		return false, err
	}
	t.IssuedAt = issuedAt
	t.Bearer = bearer
	return true, nil
}
