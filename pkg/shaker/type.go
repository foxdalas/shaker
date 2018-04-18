package shaker

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"github.com/go-redis/redis"

)

type Shaker struct {
	version string
	log     *log.Entry

	Config *Config

	bitbucketUser     string
	bitbucketPassword string

	configFile string
	stopCh    chan struct{}
	waitGroup sync.WaitGroup
	redisClient *redis.Client
	Jobs []RunJob

	RedisStorages map[string]*redis.Client
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
			Storages map[string]RedisStorage `json:"Storages"`
			Dir string `json:"dir"`
		} `json:"redis"`
	} `json:"jobs"`
	Users map[string]User `json:"Users"`
}

type RedisStorage struct {
	Host string
	Port string
}

type User struct {
	User string
	Password string
}

type Jobs struct {
	URL  string `json:"url"`
	Redis string `json:"redis"`
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
	Channel string `json:"channel"`
	Message string `json:"message"`


}

type RunJob struct {
	Name        string
	URL         string
	RedisClient	string
	Type        string
	Method      string
	Username    string
	Password    string
	Channel     string
	Message     string
	log         *log.Entry
	redisLock 	*redis.Client
	lockTimeout int
	jobFile string
	redisStorage *redis.Client
}
