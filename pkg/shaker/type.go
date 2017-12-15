package shaker

import (
	log "github.com/sirupsen/logrus"
	"time"
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

type Job struct {
	Name     string        `json:"name"`
	Url      string        `json:"url"`
	Runner   string        `json:"runner"`
	Duration time.Duration `json:"duration"`
}

type Config struct {
	Environment string `json:"environment"`
	Role        string `json:"role"`
	Storage struct {
		Redis struct {
			Memory struct {
				Host string `json:"host"`
				Port int    `json:"port"`
			} `json:"memory"`
			Pubsub struct {
				Host string `json:"host"`
				Port int    `json:"port"`
			} `json:"pubsub"`
		} `json:"redis"`
	} `json:"storage"`
	Applications []struct {
		Prefix string `json:"prefix"`
		Config string `json:"config"`
	} `json:"applications"`
	Master struct {
		Socket string `json:"socket"` //@see: http://api.zeromq.org/4-1:zmq-connect#toc2
	} `json:"master"`
	IsMaster bool  `json:"isMaster"`
	Jobs     []Job `json:"jobs"`
}
