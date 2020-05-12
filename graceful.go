package graceful

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/mylxsw/asteria/log"
)

type Graceful struct {
	lock sync.Mutex

	reloadSignals   []os.Signal
	shutdownSignals []os.Signal

	perHandlerTimeout time.Duration

	signalChan chan os.Signal

	reloadHandlers   []func()
	shutdownHandlers []func()
}

func New(reloadSignals []os.Signal, shutdownSignals []os.Signal, perHandlerTimeout time.Duration) *Graceful {
	return &Graceful{
		reloadSignals:     reloadSignals,
		shutdownSignals:   shutdownSignals,
		reloadHandlers:    make([]func(), 0),
		shutdownHandlers:  make([]func(), 0),
		perHandlerTimeout: perHandlerTimeout,
		signalChan:        make(chan os.Signal),
	}
}

func (gf *Graceful) AddReloadHandler(h func()) {
	gf.lock.Lock()
	defer gf.lock.Unlock()

	gf.reloadHandlers = append(gf.reloadHandlers, h)
}

func (gf *Graceful) AddShutdownHandler(h func()) {
	gf.lock.Lock()
	defer gf.lock.Unlock()

	gf.shutdownHandlers = append(gf.shutdownHandlers, h)
}

func (gf *Graceful) Reload() {
	log.Debug("execute reload...")
	go gf.reload()
}

func (gf *Graceful) Shutdown() {
	log.Debug("shutdown...")

	_ = gf.SignalSelf(os.Interrupt)
}

func (gf *Graceful) SignalSelf(sig os.Signal) error {
	gf.signalChan <- sig
	return nil
}

func (gf *Graceful) shutdown() {
	gf.lock.Lock()
	defer gf.lock.Unlock()

	ok := make(chan interface{}, 0)
	defer close(ok)
	for i := len(gf.shutdownHandlers) - 1; i >= 0; i-- {

		go func(handler func()) {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("execute shutdown handler failed: %s", err)
				}
				ok <- struct{}{}
			}()

			handler()
		}(gf.shutdownHandlers[i])

		select {
		case <-ok:
		case <-time.After(gf.perHandlerTimeout):
			log.Errorf("execute shutdown handler timeout")
		}
	}
}

func (gf *Graceful) reload() {
	gf.lock.Lock()
	defer gf.lock.Unlock()

	ok := make(chan interface{}, 0)
	defer close(ok)
	for i := len(gf.reloadHandlers) - 1; i >= 0; i-- {
		go func(handler func()) {
			defer func() {
				if err := recover(); err != nil {
					log.Errorf("execute reload handler failed: %s", err)
				}
				ok <- struct{}{}
			}()
			handler()
		}(gf.reloadHandlers[i])

		select {
		case <-ok:
		case <-time.After(gf.perHandlerTimeout):
			log.Errorf("execute reload handler timeout")
		}
	}
}

func (gf *Graceful) Start() error {
	signals := make([]os.Signal, 0)
	signals = append(signals, gf.reloadSignals...)
	signals = append(signals, gf.shutdownSignals...)

	signal.Notify(gf.signalChan, signals...)
	for {
		sig := <-gf.signalChan

		for _, s := range gf.shutdownSignals {
			if s == sig {
				goto FINAL
			}
		}

		for _, s := range gf.reloadSignals {
			if s == sig {
				log.Debugf("received a reload signal %s", sig.String())
				gf.reload()
				break
			}
		}
	}
FINAL:

	log.Debug("received a shutdown signal")

	gf.shutdown()

	return nil
}
