// +build windows

package graceful

import (
	"os"
	"time"
)

func NewWithDefault(perHandlerTimeout time.Duration) Graceful {
	return NewWithSignal([]os.Signal{}, []os.Signal{os.Interrupt}, perHandlerTimeout)
}
