// +build !windows

package internal

import (
	"github.com/reddec/trusted-cgi/types"
	"os/exec"
	"syscall"
)

func SetCreds(cmd *exec.Cmd, creds *types.Credential) {
	if creds == nil {
		return
	}
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: uint32(creds.User),
		Gid: uint32(creds.Group),
	}
}
