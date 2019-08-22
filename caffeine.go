package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/shu-go/gli"
)

var (
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	setThreadExecutionState = kernel32.NewProc("SetThreadExecutionState")
)

const (
	ES_CONTINUOUS        = 0x80000000
	ES_SYSTEM_REQUIRED   = 0x00000001
	ES_DISPLAY_REQUIRED  = 0x00000002
	ES_USER_PRESENT      = 0x00000004
	ES_AWAYMODE_REQUIRED = 0x00000040
)

type globalCmd struct {
}

func (c globalCmd) Run() error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	setThreadExecutionState.Call(uintptr(ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED | ES_CONTINUOUS))

	fmt.Println("Press Ctrl+C to stop.")

	/*
		loop:
			for {
				select {
				case <-signalChan:
					break loop
				case <-time.After(30 * time.Second):
					setThreadExecutionState.Call(uintptr(ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED))
				}
			}
	*/
	<-signalChan

	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS))

	return nil
}

func main() {
	app := gli.NewWith(&globalCmd{})
	app.Name = "caffeine"
	app.Desc = "keep waking Windows up"
	app.Version = "0.1.0"
	app.Usage = "caffeine"
	app.Copyright = "(C) 2019 Shuhei Kubota"
	app.Run(os.Args)
}
