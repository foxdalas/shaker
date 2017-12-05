package shaker

import (
	log "github.com/sirupsen/logrus"
)

type Shaker struct {
	config     Config
	version    string
	log        *log.Entry
	stopCh     chan struct{}
	Jobs       []RunJob
	jobRunners map[string]JobRunner
}

type CronData []struct {
	Name string `json:"name"`
	Cron string `json:"cron"`
	URI  string `json:"uri"`
}

type RunJob struct {
	Name string
	URL  string
	log  *log.Entry
}

type JobRunner interface {
	Init(config Config, log *log.Entry) error
	Run(job Job) error
	Schedule(job Job)
}
