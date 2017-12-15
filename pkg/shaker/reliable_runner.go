package shaker

import (
	log "github.com/sirupsen/logrus"
	"time"
)

type ReliableRunner struct {
	log *log.Entry
}

func (runner *ReliableRunner) Init(config Config, log *log.Entry) error {
	return nil
}

func (runner ReliableRunner) Run(job Job) error {
	return Fetch(job.Url)
}

func (runner *ReliableRunner) Schedule(job Job) {
	ticker := time.NewTicker(job.Duration)
	go func() {
		for range ticker.C {
			err := runner.Run(job)
			if err != nil {
				runner.log.Errorf("Error: %s", err)
			}
		}
	}()
}
