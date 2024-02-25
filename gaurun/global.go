package gaurun

import (
	"github.com/mercari/gaurun/gcm"

	firebase "firebase.google.com/go"

	"go.uber.org/zap"
)

var (
	// Toml configuration for Gaurun
	ConfGaurun ConfToml
	// push notification Queue
	QueueNotification chan RequestGaurunNotification
	// Stat for Gaurun
	StatGaurun StatApp
	// http client for APNs and GCM/FCM
	APNSClient  APNsClient
	GCMClient   *gcm.Client
	FirebaseApp *firebase.App
	// access and error logger
	LogAccess *zap.Logger
	LogError  *zap.Logger
	// sequence ID for numbering push
	SeqID uint64
)
