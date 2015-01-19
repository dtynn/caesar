package gracefuldown

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"

	"github.com/qiniu/log"
)

var (
	GracefulDownCode = 497
	gdDefaultTimeout = 10 * time.Second
	gd               = &gracefulDowner{
		timeout: gdDefaultTimeout,
	}
)

type gracefulDowner struct {
	running  int32
	timeout  time.Duration
	shutdown int32
	handling int32
}

func (this *gracefulDowner) run() {
	atomic.StoreInt32(&gd.running, 1)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

WAIT:
	for {
		select {
		case <-c:
			break WAIT
		}
	}
	log.Debug("shutting down")
	atomic.StoreInt32(&gd.shutdown, 1)

	timer := time.NewTimer(this.timeout)

	for {
		select {
		case <-timer.C:
			log.Warnf("graceful down timeout with %d requests", atomic.LoadInt32(&gd.handling))
			os.Exit(-1)
		default:
			if gd.handling <= 0 {
				log.Debug("graceful down done")
				os.Exit(0)
			}
		}
	}

}

func SetGracefulDownTimeout(timeout time.Duration) {
	if atomic.LoadInt32(&gd.running) == 0 {
		gd.timeout = timeout
	}
}

func Run() {
	go gd.run()
}

func GracefulBefore(w http.ResponseWriter, r *http.Request) (int, error) {
	atomic.AddInt32(&gd.handling, 1)

	if atomic.LoadInt32(&gd.shutdown) == 1 {
		log.Debug("request blocked")
		return GracefulDownCode, fmt.Errorf("this server is unavailable")
	}
	return 0, nil
}

func GracefulAfter(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt32(&gd.handling, -1)
	return
}
