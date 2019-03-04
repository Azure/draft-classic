// +build windows

// NOTICE: This implementation comes from logrus, unfortunately they
// do not expose a public interface to call it.
//   https://github.com/sirupsen/logrus/blob/master/terminal_check_notappengine.go
//   https://github.com/sirupsen/logrus/blob/master/terminal_windows.go

package cmdline

import (
	"io"
	"os"
	"syscall"

	sequences "github.com/konsorten/go-windows-terminal-sequences"
)

// initTerminal enables ANSI color escape on windows. Usually, this is done by logrus,
// but only after the first log message. So instead we take care of this ourselves.
func initTerminal(w io.Writer) {
	if f, ok := w.(*os.File); ok {
		sequences.EnableVirtualTerminalProcessing(syscall.Handle(f.Fd()), true)
	}
}

func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		var mode uint32
		err := syscall.GetConsoleMode(syscall.Handle(f.Fd()), &mode)
		return err == nil
	}
	return false
}
