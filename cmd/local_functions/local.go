package main

import (
	"context"
	"log"
	"os"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"

	"github.com/nohana/gaurun/functions/gaurun"
)

func main() {

	// Set the PUBSUB_EMULATOR_HOST environment variable.
	os.Setenv("PUBSUB_EMULATOR_HOST", "localhost:8043")

	ctx := context.Background()

	gaurun.InitGaurun()

	pushMethod := gaurun.PushFromEvent

	// 実行する関数の登録
	if err := funcframework.RegisterCloudEventFunctionContext(ctx, "/", pushMethod); err != nil {
		log.Fatalf("err := funcframework.RegisterCloudEventFunctionContext: %v\n", err)
	}

	port := "8080"
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v\n", err)
	}
}
