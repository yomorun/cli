// +build !windows

package cmd

import (
	"os"
	"os/signal"
	"syscall"
)

func setupShutdownNotify(sigCh chan os.Signal) {
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
}
