package gaurun

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"google.golang.org/api/option"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"

	"github.com/nohana/gaurun/buford/token"
	"github.com/nohana/gaurun/gcm"
)

type SafeMessagingClient struct {
	client *messaging.Client
	mu     sync.Mutex
}

func (s *SafeMessagingClient) Send(ctx context.Context, message *messaging.Message) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.client.Send(ctx, message)
}

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
	//transportを自作しWithHttpClientするとAPNsのエラーが出るようになるため、v1ライブラリを使う場合はHttpClientの動作を指定しない
	//transport := &http.Transport{
	//	MaxIdleConnsPerHost: ConfGaurun.Android.KeepAliveConns,
	//	DialContext: (&net.Dialer{
	//		Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
	//		KeepAlive: time.Duration(keepAliveInterval(ConfGaurun.Android.KeepAliveTimeout)) * time.Second,
	//	}).DialContext,
	//	IdleConnTimeout: time.Duration(ConfGaurun.Android.KeepAliveTimeout) * time.Second,
	//}
	//
	//client := &http.Client{
	//	Transport: transport,
	//	Timeout:   time.Duration(ConfGaurun.Android.Timeout) * time.Second,
	//}

	opts := make([]option.ClientOption, 1)
	opts[0] = InitSaOption()
	//opts[1] = option.WithHTTPClient(client)

	var err error

	ctx := context.Background()
	FirebaseApp, err = firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return err
	}

	msgClient, err := FirebaseApp.Messaging(ctx)
	if err != nil {
		return err
	}

	FcmV1Client = &SafeMessagingClient{
		client: msgClient,
		mu:     sync.Mutex{},
	}

	if err != nil {
		return err
	}

	return nil
}

func InitSaOption() option.ClientOption {
	var saOpt option.ClientOption
	if len(ConfGaurun.Android.CredentialsJSONBase64) > 0 {
		// Base64デコード
		decodedSaJsonBytes, err := base64.StdEncoding.DecodeString(ConfGaurun.Android.CredentialsJSONBase64)
		if err != nil {
			fmt.Println("config load sa base64 json decode error", err)
		} else {
			// デコードされた結果を文字列に変換
			saJson := string(decodedSaJsonBytes)
			saOpt = option.WithCredentialsJSON([]byte(saJson))
		}
	} else {
		saOpt = option.WithCredentialsFile(ConfGaurun.Android.CredentialsFile)
	}
	return saOpt
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
