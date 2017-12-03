package shaker

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type Shaker struct {
	version string
	log     *log.Entry

	bitbucketUser     string
	bitbucketPassword string

	configFile string

	stopCh    chan struct{}
	waitGroup sync.WaitGroup

	Jobs []RunJob
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
	}
	IsMaster bool `json:"isMaster"`
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

type Job struct {
	Name string
	Url  string
	Runner string
	Duration time.Duration
}

type JobRunner interface {
	Init(config Config, log *log.Entry) error
	Run(job Job) error
	Schedule(job Job)
}
