package shaker

import (
	"sync"
	"time"

	"github.com/bsm/redis-lock"
	"github.com/go-redis/redis"
	"github.com/nlopes/slack"
	log "github.com/sirupsen/logrus"
)

//Shaker struct
type Shaker struct {
	version    string
	log        *log.Entry
	config     config
	connectors struct {
		redisStorages map[string]*redis.Client
		slackConfig   slackConfig
	}
	jobs      []RunJob
	stopCh    chan struct{}
	waitGroup sync.WaitGroup
}

type config struct {
	Environment string `json:"environment"`
	Role        string `json:"role"`
	Storage     struct {
		Redis struct {
			Default struct {
				Host     string `json:"host"`
				Port     string `json:"port"`
				Password string `json:"password"`
			} `json:"default"`
			Memory struct {
				Host     string `json:"host"`
				Port     string `json:"port"`
				Password string `json:"password"`
			} `json:"memory"`
			Pubsub struct {
				Host     string `json:"host"`
				Port     string `json:"port"`
				Password string `json:"password"`
			} `json:"pubsub"`
		} `json:"redis"`
	} `json:"storage"`
	Slack struct {
		Enabled bool   `json:"enabled"`
		Channel string `json:"channel"`
		Token   string `json:"token"`
	} `json:"slack"`
	Jobs struct {
		HTTP struct {
			Dir string `json:"dir"`
		} `json:"http"`
		Redis struct {
			Storages map[string]redisStorage `json:"Storages"`
			Dir      string                  `json:"dir"`
		} `json:"redis"`
	} `json:"jobs"`
	Watch struct {
		Dir string `json:"dir"`
	} `json:"watch"`
	Users map[string]user `json:"Users"`
}

type redisStorage struct {
	Host string
	Port string
}

type user struct {
	User     string
	Password string
}

type jobs struct {
	URL string `json:"url"`
	//Redis string `json:"redis"`
	Jobs []job `json:"jobs"`
}

type job struct {
	Name        string `json:"name"`
	Cron        string `json:"cron"`
	URI         string `json:"uri"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	LockTimeout int    `json:"lock"`
	Method      string `json:"method"`
	Channel     string `json:"channel"`
	Message     string `json:"message"`
	Timeout     int    `json:"timeout"`
}

//RunJob structure for store job parameters
type request struct {
	name        string //Cronjob Name
	url         string //HTTP URL
	method      string //TODO: Cleanup after refactoring
	requestType string //Request type GET/POST/Publish
	username    string //HTTP Basic Auth username
	password    string //HTTP Basic Auth password
	channel     string //Redis Channel
	message     string //Redis Message
	timeout     time.Duration //Request timeout
}

type clients struct {
	redisClient  string
	redisStorage *redis.Client
	slackClient  slackConfig
}

type RunJob struct {
	log  *log.Entry
	lock *lock.Locker

	request request
	clients *clients
}

type slackConfig struct {
	enabled bool
	client  *slack.Client
	channel string
}
