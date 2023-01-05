package internal

import (
	"os/exec"
	"syscall"

	"github.com/reddec/trusted-cgi/types"
)

func Reap(pid int) {
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}

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
	cmd.SysProcAttr.Setpgid = true
	cmd.SysProcAttr.Pdeathsig = syscall.SIGTERM
}

const Shell = "/bin/sh"
