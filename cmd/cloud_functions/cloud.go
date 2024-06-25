package main

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"github.com/nohana/gaurun/functions/gaurun"
)

func Init() {
	gaurun.InitGaurun()
	pushMethod := gaurun.PushFromEvent
	functions.CloudEvent("PushFromEvent", pushMethod)
}
