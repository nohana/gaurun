package gaurun

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"net"
	"net/http"
	"time"

	firebase "firebase.google.com/go"

	"google.golang.org/api/option"

	"github.com/mercari/gaurun/buford/token"
	"github.com/mercari/gaurun/gcm"
)

func keepAliveInterval(keepAliveTimeout int) int {
	const minInterval = 30
	const maxInterval = 90
	if keepAliveTimeout <= minInterval {
		return keepAliveTimeout
	}
	result := keepAliveTimeout / 3
	if result < minInterval {
		return minInterval
	}
	if result > maxInterval {
		return maxInterval
	}
	return result
}

// InitGCMClient initializes GCMClient which is globally declared.
func InitGCMClient() error {
	var err error
	GCMClient, err = gcm.NewClient(gcm.FCMSendEndpoint, ConfGaurun.Android.ApiKey)
	if err != nil {
		return err
	}

	transport := &http.Transport{
		MaxIdleConnsPerHost: ConfGaurun.Android.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
			KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.Android.KeepAliveTimeout)) * time.Second,
		}).Dial,
		IdleConnTimeout: time.Duration(ConfGaurun.Android.KeepAliveTimeout) * time.Second,
	}

	GCMClient.Http = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
	}

	return nil
}

func InitFirebaseAppForFcmV1() error {
	transport := &http.Transport{
		MaxIdleConnsPerHost: ConfGaurun.Android.KeepAliveConns,
		Dial: (&net.Dialer{
			Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
			KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.Android.KeepAliveTimeout)) * time.Second,
		}).Dial,
		IdleConnTimeout: time.Duration(ConfGaurun.Android.KeepAliveTimeout) * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
	}

	opts := make([]option.ClientOption, 2)
	opts[0] = option.WithCredentialsFile(ConfGaurun.Android.CredentialsFile)
	opts[1] = option.WithHTTPClient(client)

	// if ConfGaurun.Android.Project is empty string, it is acquired from the contents of ConfGaurun.Android.CredentialsFile
	config := &firebase.Config{ProjectID: ConfGaurun.Android.Project}

	var err error

	FirebaseApp, err = firebase.NewApp(context.Background(), config, opts...)
	if err != nil {
		return err
	}

	return nil
}

func InitAPNSClient() error {
	var err error
	if ConfGaurun.Ios.IsCertificateBasedProvider() {
		APNSClient, err = NewApnsClientHttp2(
			ConfGaurun.Ios.PemCertPath,
			ConfGaurun.Ios.PemKeyPath,
			ConfGaurun.Ios.PemKeyPassphrase,
		)
	} else if ConfGaurun.Ios.IsTokenBasedProvider() {
		var authKey *ecdsa.PrivateKey
		authKey, err = token.AuthKeyFromFile(ConfGaurun.Ios.TokenAuthKeyPath)
		if err != nil {
			return err
		}
		APNSClient, err = NewApnsClientHttp2ForToken(
			authKey,
			ConfGaurun.Ios.TokenAuthKeyID,
			ConfGaurun.Ios.TokenAuthTeamID,
		)
	} else {
		return fmt.Errorf("should be specify Token-based provider or Certificate-based provider")
	}
	if err != nil {
		return err
	}
	return nil
}
