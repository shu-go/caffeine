package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/shu-go/gli/v2"
	"github.com/shu-go/shortcut"
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
	esContinuous      = 0x80000000
	esSystemRequired  = 0x00000001
	esDisplayRequired = 0x00000002
	//unused
	//esUserPresent      = 0x00000004
	//esAwaymodeRequired = 0x00000040
)

type globalCmd struct {
	DoRun runCmd `cli:"run,r"`

	Timeout time.Duration `cli:"t,timeout=DURATION" default:"-1s"`
	Dest    string        `cli:"dest,d=PATH_TO_LNK" defdesc:"(default: dir of caffeine.exe)"`
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

	binpath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("creating shortcut, executable: %v", err)
	}
	target := strings.Join(args, " ")
	lnkpath := filepath.Join(c.Dest, filepath.Base(target))

	s := shortcut.New(binpath)
	var ss *shortcut.Shortcut
	if strings.HasSuffix(strings.ToLower(target), ".lnk") {
		ss, err = shortcut.Open(target)
		if err != nil {
			ss = nil
		}
	}
	if ss != nil {
		fmt.Printf("%v\n", *ss)
		*s = *ss
		s.Arguments = `run "` + ss.TargetPath + `"`
		s.TargetPath = binpath
	} else {
		s.Arguments = `run "` + target + `"`
		s.IconLocation = target + ",0"
		lnkpath = lnkpath[:len(lnkpath)-len(filepath.Ext(lnkpath))]
	}
	s.WindowStyle = 7 // min
	fmt.Printf("%v\n", *s)

	return s.Save(lnkpath)
}

func (c globalCmd) runStandalone() error {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	setThreadExecutionState.Call(uintptr(esSystemRequired | esDisplayRequired | esContinuous))

	fmt.Println("Press Ctrl+C to stop.")

	/*
		loop:
			for {
				select {
				case <-signalChan:
					break loop
				case <-time.After(30 * time.Second):
					setThreadExecutionState.Call(uintptr(esSystemRequired | esDisplayRequired))
				}
			}
	*/
	if c.Timeout > time.Duration(0) {
		fmt.Println("timeout: " + c.Timeout.String())

		select {
		case <-time.After(c.Timeout):
		case <-signalChan:
		}
	} else {
		<-signalChan
	}

	setThreadExecutionState.Call(uintptr(esContinuous))

	return nil
}

func main() {
	app := gli.NewWith(&globalCmd{})
	app.Name = "caffeine"
	app.Desc = "keep waking Windows up"
	app.Version = Version
	app.Usage = `RUN AS ADMINISTRATOR

# standalone mode
    # start
    > ./caffeine

    # termination
    Ctrl+C

# app mode
    # start
    > ./caffeine run PATH_TO_EXE

    # termination
    terminate the app

# shortcut-creation mode
    # preparation
    > ./caffeine PATH_TO_EXE

    # start
    start the shortcut

    # termination
    terminate the app`
	app.Copyright = "(C) 2019 Shuhei Kubota"
	app.Run(os.Args)
}
