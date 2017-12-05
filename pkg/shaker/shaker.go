package shaker

import (
	//log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
	//"io/ioutil"
	//"net/http"
	//"github.com/gin-gonic/gin"
	//"time"
)

func New(config Config, version string) *Shaker {
	logger := CreateLogger()
	sh := &Shaker{
		config:     config,
		jobRunners: CreateRunnerContainer(config, logger),
		version:    version,
		log:        logger,
		stopCh:     make(chan struct{}),
	}
	return sh
}

func (sh *Shaker) Start() {
	sh.log.Infof("Shaker %s starting", sh.version)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-c
		logger := sh.log.WithField("signal", s.String())
		logger.Debug("received signal")
		sh.Stop()
	}()

	for _, job := range sh.config.Jobs {
		go sh.jobRunners[job.Runner].Schedule(job)
	}
}

func (sh *Shaker) Version() string {
	return sh.version
}

func (sh *Shaker) Stop() {
	sh.log.Info("shutting things down")
	close(sh.stopCh)
}

//func (e RunJob) Run() {
//	start := time.Now()
//	resp, err := http.Get(e.URL)
//	if err != nil {
//		e.log.Error(err)
//	}
//	defer resp.Body.Close()
//	elapsed := time.Since(start).Seconds()
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		e.log.Errorf("Error: %s", err)
//	}
//	e.log = log.WithFields(log.Fields{
//		"context":       "shaker",
//		"response":      resp.StatusCode,
//		"response_time": elapsed,
//		"request":       e.URL,
//	})
//	e.log.Info(string(body))
//}
