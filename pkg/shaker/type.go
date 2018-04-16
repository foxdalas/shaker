package shaker

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"github.com/go-redis/redis"

)

type Shaker struct {
	version string
	log     *log.Entry

	bitbucketUser     string
	bitbucketPassword string

	configFile string
	stopCh    chan struct{}
	waitGroup sync.WaitGroup
	redisClient *redis.Client
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
	Redis struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Password string `json:"password"`
	} `json:"redis"`
	Jobs struct {
		Http struct {
			Dir string `json:"dir"`
		} `json:"http"`
		Redis struct {
			Dir string `json:"dir"`
		} `json:"redis"`
	} `json:"jobs"`
	Users map[string]User `json:"Users"`
}


type User struct {
	User string
	Password string
}

type HTTPJobs struct {
	URL  string `json:"url"`
	Jobs []Job  `json:"jobs"`
}

type Job struct {
	Name        string `json:"name"`
	Cron        string `json:"cron"`
	URI         string `json:"uri"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	LockTimeout int    `json:"lock"`
	Method      string `json:"method"`
}

type RunJob struct {
	Name        string
	URL         string
	Type        string
	Method      string
	Username    string
	Password    string
	log         *log.Entry
	redisClient *redis.Client
	lockTimeout int
	jobFile string
}
