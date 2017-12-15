package shaker

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"github.com/go-redis/redis"
)

type Shaker struct {

	version string
	log     *log.Entry

	bitbucketUser string
	bitbucketPassword string

	configFile  string

	stopCh    chan struct{}
	waitGroup sync.WaitGroup

	redisClient *redis.Client

	Jobs []RunJob
}


type Config struct {
	Environment string `json:"environment"`
	Role        string `json:"role"`
	Storage     struct {
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
	Redis struct {
		Host string `json:"host"`
		Port string `json:"port"`
		Password string `json:"password"`
	} `json:"redis"`
}

type CronData []struct {
	Name string `json:"name"`
	Cron string `json:"cron"`
	URI  string `json:"uri"`
	LockTimeout int `json:"lock"`
}

type RunJob struct {
	Name string
	URL string
	log *log.Entry
	redisClient *redis.Client
	lockTimeout int
}