// +build !windows

package main

import (
	"golang.org/x/crypto/ssh/terminal"
	"syscall"
)

func AskPass() ([]byte, error) {
	return terminal.ReadPassword(syscall.Stdin)
}
