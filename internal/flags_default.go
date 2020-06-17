//+build !linux

package internal

import "os/exec"

func SetFlags(cmd *exec.Cmd) {}
