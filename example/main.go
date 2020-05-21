package main

import (
	"time"

	"github.com/mylxsw/asteria/log"
	"github.com/mylxsw/graceful"
)

type Service struct {
	stop chan interface{}
	gf   graceful.Graceful
}

func (s *Service) Init(gf graceful.Graceful) {
	s.stop = make(chan interface{}, 0)
	s.gf = gf
}

func (s *Service) Start() {
	startTs := time.Now()
	for {
		select {
		case <-s.stop:
			log.Debug("receive stop signal, exit")
		default:
			log.Debug("Hello, world")
			time.Sleep(1 * time.Second)
			if time.Now().Sub(startTs) > 5*time.Second {
				s.gf.Shutdown()
			}
		}
	}
}

func (s *Service) Stop() {
	//time.Sleep(5 * time.Second)
	s.stop <- struct{}{}
}

func main() {
	gf := graceful.NewWithDefault(3 * time.Second)

	s := &Service{}
	s.Init(gf)

	gf.AddShutdownHandler(s.Stop)
	go s.Start()

	if err := gf.Start(); err != nil {
		log.Error(err)
	}
}
