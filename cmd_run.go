package main

import (
	"errors"
	"os"
	"os/exec"
)

type runCmd struct {
}

func (c *runCmd) Run(args []string) error {
	if len(args) == 0 {
		return errors.New("no path")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	setThreadExecutionState.Call(uintptr(ES_SYSTEM_REQUIRED | ES_DISPLAY_REQUIRED | ES_CONTINUOUS))

	err := cmd.Run()

	setThreadExecutionState.Call(uintptr(ES_CONTINUOUS))

	if err != nil {
		return err
	}

	return nil
}
