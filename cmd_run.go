package main

import (
	"errors"
	"os"
	"os/exec"
)

type runCmd struct {
	_ struct{} `help:"wakes Windows up, then runs passed app and waits for it's termination, then allow Windows sleep"`
}

func (c *runCmd) Run(args []string) error {
	if len(args) == 0 {
		return errors.New("no path")
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	setThreadExecutionState.Call(uintptr(esSystemRequired | esDisplayRequired | esContinuous))

	err := cmd.Run()

	setThreadExecutionState.Call(uintptr(esContinuous))

	if err != nil {
		return err
	}

	return nil
}
