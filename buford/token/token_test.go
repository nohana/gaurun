package token_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/nohana/gaurun/buford/token"
	"github.com/stretchr/testify/assert"
)

// AuthToken

func TestValidTokenFromP8File(t *testing.T) {
	_, err := token.AuthKeyFromFile("testdata/authkey-valid.p8")
	assert.NoError(t, err)
}

func TestValidTokenFromP8Bytes(t *testing.T) {
	bytes, _ := os.ReadFile("testdata/authkey-valid.p8")
	_, err := token.AuthKeyFromBytes(bytes)
	assert.NoError(t, err)
}

func TestNoSuchFileP8File(t *testing.T) {
	keyToken, err := token.AuthKeyFromFile("")
	assert.Equal(t, errors.New("open : no such file or directory").Error(), err.Error())
	assert.Nil(t, keyToken)
}

func TestInvalidP8File(t *testing.T) {
	_, err := token.AuthKeyFromFile("testdata/authkey-invalid.p8")
	assert.Error(t, err)
}

func TestInvalidPKCS8P8File(t *testing.T) {
	_, err := token.AuthKeyFromFile("testdata/authkey-invalid-pkcs8.p8")
	assert.Error(t, err)
}

func TestInvalidECDSAP8File(t *testing.T) {
	_, err := token.AuthKeyFromFile("testdata/authkey-invalid-ecdsa.p8")
	assert.Error(t, err)
}

// Expiry & Generation

func TestExpired(t *testing.T) {
	keyToken := &token.Token{}
	assert.True(t, keyToken.Expired())
}

func TestNotExpired(t *testing.T) {
	keyToken := &token.Token{
		IssuedAt: time.Now().Unix(),
	}
	assert.False(t, keyToken.Expired())
}

func TestExpiresBeforeAnHour(t *testing.T) {
	keyToken := &token.Token{
		IssuedAt: time.Now().Add(-50 * time.Minute).Unix(),
	}
	assert.True(t, keyToken.Expired())
}

func TestGenerateBearerIfExpired(t *testing.T) {
	authKey, _ := token.AuthKeyFromFile("testdata/authkey-valid.p8")
	keyToken := &token.Token{
		AuthKey: authKey,
	}
	keyToken.GenerateBearerIfExpired()
	assert.Equal(t, time.Now().Unix(), keyToken.IssuedAt)
}

func TestGenerateWithNoAuthKey(t *testing.T) {
	keyToken := &token.Token{}
	isSuccess, err := keyToken.Generate()
	assert.False(t, isSuccess)
	assert.Error(t, err)
}

func TestGenerateWithInvalidAuthKey(t *testing.T) {
	pubkeyCurve := elliptic.P521()
	privatekey, _ := ecdsa.GenerateKey(pubkeyCurve, rand.Reader)
	keyToken := &token.Token{
		AuthKey: privatekey,
	}
	isSuccess, err := keyToken.Generate()
	assert.False(t, isSuccess)
	assert.Error(t, err)
}
