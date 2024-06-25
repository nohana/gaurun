package gaurun

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/nohana/gaurun/buford/token"
	"github.com/nohana/gaurun/gaurun"
)

func InitGaurun() {
	confPath := flag.String("c", "", "configuration file path for gaurun")
	listenPort := flag.String("p", "", "port number or unix socket path")
	workerNum := flag.Int64("w", 0, "number of workers for push notification")
	queueNum := flag.Int64("q", 0, "size of internal queue for push notification")
	flag.Parse()

	// set default parameters
	gaurun.ConfGaurun = gaurun.BuildDefaultConf()

	// load configuration
	var conf gaurun.ConfToml
	conf, err := gaurun.LoadConfFromEnv(gaurun.ConfGaurun)
	fmt.Println("try config load from env")
	if err != nil {
		fmt.Printf("failed load from env: %v", err)
		conf, err = gaurun.LoadConf(gaurun.ConfGaurun, *confPath)
		if err != nil {
			gaurun.LogSetupFatal(err)
		}
	}
	gaurun.ConfGaurun = conf

	// overwrite if port is specified by flags
	if *listenPort != "" {
		gaurun.ConfGaurun.Core.Port = *listenPort
	}

	// overwrite if workerNum is specified by flags
	if *workerNum > 0 {
		gaurun.ConfGaurun.Core.WorkerNum = *workerNum
	}

	// overwrite if queueNum is specified by flags
	if *queueNum > 0 {
		gaurun.ConfGaurun.Core.QueueNum = *queueNum
	}

	// set logger
	accessLogger, _, err := gaurun.InitLog(gaurun.ConfGaurun.Log.AccessLog, "info")
	if err != nil {
		gaurun.LogSetupFatal(err)
	}
	errorLogger, _, err := gaurun.InitLog(gaurun.ConfGaurun.Log.ErrorLog, gaurun.ConfGaurun.Log.Level)
	if err != nil {
		gaurun.LogSetupFatal(err)
	}

	gaurun.LogAccess = accessLogger
	gaurun.LogError = errorLogger

	if !gaurun.ConfGaurun.Ios.Enabled && !gaurun.ConfGaurun.Android.Enabled {
		gaurun.LogSetupFatal(fmt.Errorf("no platform has been enabled"))
	}

	if gaurun.ConfGaurun.Ios.Enabled {
		if gaurun.ConfGaurun.Ios.IsCertificateBasedProvider() && gaurun.ConfGaurun.Ios.IsTokenBasedProvider() {
			gaurun.LogSetupFatal(fmt.Errorf("you can use only one of certificate-based provider or token-based provider connection trust"))
		}

		if gaurun.ConfGaurun.Ios.IsCertificateBasedProvider() {
			_, err = os.ReadFile(gaurun.ConfGaurun.Ios.PemCertPath)
			if err != nil {
				gaurun.LogSetupFatal(fmt.Errorf("the certification file for iOS was not found"))
			}

			_, err = os.ReadFile(gaurun.ConfGaurun.Ios.PemKeyPath)
			if err != nil {
				gaurun.LogSetupFatal(fmt.Errorf("the key file for iOS was not found"))
			}
		} else if gaurun.ConfGaurun.Ios.IsTokenBasedProvider() {
			_, err = token.AuthKeyFromConfig(gaurun.ConfGaurun.Ios.TokenAuthKeyPath, gaurun.ConfGaurun.Ios.TokenAuthKeyBase64)
			if err != nil {
				gaurun.LogSetupFatal(fmt.Errorf("the auth key file for iOS was not loading: %v", err))
			}
		} else {
			gaurun.LogSetupFatal(fmt.Errorf("the key file or APNsAuthKey file for iOS was not found"))
		}
	}

	if gaurun.ConfGaurun.Android.Enabled {
		if gaurun.ConfGaurun.Android.ApiKey == "" {
			gaurun.LogSetupFatal(fmt.Errorf("the APIKey for Android cannot be empty"))
		}
	}

	sigHUPChan := make(chan os.Signal, 1)
	signal.Notify(sigHUPChan, syscall.SIGHUP)

	if gaurun.ConfGaurun.Android.Enabled {
		if err := gaurun.InitGCMClient(); err != nil {
			gaurun.LogSetupFatal(fmt.Errorf("failed to init gcm/fcm client: %v", err))
		}
		if gaurun.ConfGaurun.Android.UseV1 {
			if err := gaurun.InitFirebaseAppForFcmV1(); err != nil {
				gaurun.LogSetupFatal(fmt.Errorf("failed to init fcm v1 firebase messaging client: %v", err))
			}
		}
	}

	if gaurun.ConfGaurun.Ios.Enabled {
		if err := gaurun.InitAPNSClient(); err != nil {
			gaurun.LogSetupFatal(fmt.Errorf("failed to init http client for APNs: %v", err))
		}
	}

	gaurun.InitStat()
	gaurun.StartPushWorkers(gaurun.ConfGaurun.Core.WorkerNum, gaurun.ConfGaurun.Core.QueueNum)

	// Start a goroutine to log number of job queue.
	go func() {
		for {
			queue := len(gaurun.QueueNotification)
			if queue == 0 {
				break
			}

			gaurun.LogError.Info(fmt.Sprintf("wait until queue is empty. Current queue len: %d", queue))
			time.Sleep(1 * time.Second)
		}
	}()

	// Block until all pusher worker job is done.
	gaurun.PusherWg.Wait()

	gaurun.LogError.Info("successfully shutdown")
}

func PushFromEvent(ctx context.Context, e event.Event) error {
	err := gaurun.PushNotificationFromPubSub(ctx, e.Data())
	if err != nil {
		return fmt.Errorf("message:Push failed error:%s pubsub_id:%s", err, e.ID())
	}

	fmt.Printf("message:Push succeeded:pubsub_id:%s", e.ID())

	return nil
}
