package zaplogger

import (
	"github.com/bpcoder16/Chestnut/contrib/file/standard"
	"path"
	"syscall"
)

// hookStderr 劫持 Stderr
func hookStderr(logDir string) {
	fh := standard.NewWriter(path.Join(logDir, "std", "stderr.log"))
	err := syscall.Dup2(int(fh.Fd()), 2)
	if err != nil {
		panic("stderr.log syscall.Dup2 failed: " + err.Error())
	}
}

// hookStdout 劫持 Stdout
func hookStdout(logDir string) {
	fh := standard.NewWriter(path.Join(logDir, "std", "stdout.log"))
	err := syscall.Dup2(int(fh.Fd()), 1)
	if err != nil {
		panic("stdout.log syscall.Dup2 failed: " + err.Error())
	}
}
