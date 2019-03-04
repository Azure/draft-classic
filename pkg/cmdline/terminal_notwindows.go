// +build !windows

// NOTICE: This implementation comes from logrus, unfortunately they
// do not expose a public interface to call it.
//   https://github.com/sirupsen/logrus/blob/master/terminal_check_notappengine.go
//   https://github.com/sirupsen/logrus/blob/master/terminal_windows.go

package cmdline

import (
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

// initTerminal enables ANSI color escape sequences. On UNIX, they are always enabled.
func initTerminal(_ io.Writer) {
}

func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return terminal.IsTerminal(int(f.Fd()))
	}
	return false
}
