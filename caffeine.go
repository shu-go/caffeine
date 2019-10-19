package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/shu-go/gli"
)

// Version is app version
var Version string

func init() {
	if Version == "" {
		Version = "dev-" + time.Now().Format("20060102")
	}
}

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
	DoRun runCmd `cli:"run,r"`

	Dest string `cli:"dest,d=PATH_TO_LNK (default: dir of caffeine.exe)"`
}

func (c *globalCmd) Before(args []string) {
	if len(args) > 0 && c.Dest == "" {
		if exepath, err := os.Executable(); err == nil {
			c.Dest = filepath.Dir(exepath)
		}
	}
}

func (c globalCmd) Run(args []string) error {
	if len(args) == 0 {
		return c.runStandalone()
	}

	src, err := os.Executable()
	if err != nil {
		return fmt.Errorf("creating shortcut, executable: %v", err)
	}
	target := strings.Join(args, " ")
	dst := filepath.Join(c.Dest, filepath.Base(target)) + ".lnk"

	return createShortcut(src, `run "`+target+`"`, dst, 7, target+",0")
}

func (c globalCmd) runStandalone() error {
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
	app.Version = Version
	app.Usage = "caffeine"
	app.Copyright = "(C) 2019 Shuhei Kubota"
	app.Run(os.Args)
}
