// +build !windows

package graceful

import (
	"os"
	"syscall"
	"time"
)

func NewWithDefault(perHandlerTimeout time.Duration) *Graceful {
	return New([]os.Signal{syscall.SIGUSR2}, []os.Signal{os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP}, perHandlerTimeout)
}
