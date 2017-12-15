package shaker

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"strings"
	"os"
	"github.com/foxdalas/shaker/pkg/shaker_const"
	"os/signal"
	"syscall"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"encoding/json"
	"errors"
	"github.com/bamzi/jobrunner"
	"net/http"
	"github.com/gin-gonic/gin"
	"time"
	"github.com/go-redis/redis"
	"github.com/bsm/redis-lock"
	"crypto/md5"
	"encoding/hex"
)

var _ shaker.Shaker = &Shaker{}

func New(version string) *Shaker {
	return &Shaker{
		version:   version,
		log:       MakeLog(),
		stopCh:    make(chan struct{}),
		waitGroup: sync.WaitGroup{},
	}
}

func (sh *Shaker) Init() {
	sh.Log().Infof("Shaker %s starting", sh.version)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-c
		logger := sh.Log().WithField("signal", s.String())
		logger.Debug("received signal")
		sh.Stop()
	}()

	err := sh.params()
	if err != nil {
		sh.Log().Fatal(err)
	}
	sh.getCronList(sh.getShakerConfig(sh.configFile))


}

func (sh *Shaker) params() error {

	sh.bitbucketUser = os.Getenv("BITBUCKET_USERNAME")
	if len(sh.bitbucketUser) == 0 {
		return errors.New("Please provide the secret key via environment variable BITBUCKET_USERNAME")
	}

	sh.bitbucketPassword = os.Getenv("BITBUCKET_PASSWORD")
	if len(sh.bitbucketPassword) == 0 {
		return errors.New("Please provide the secret key via environment variable BITBUCKET_PASSWORD")
	}

	sh.configFile = os.Getenv("CONFIG")
	if len(sh.configFile) == 0 {
		return errors.New("Please provide the secret key via environment variable CONFIG")
	}

	return nil
}

func MakeLog() *log.Entry {
	logtype := strings.ToLower(os.Getenv("LOG_TYPE"))
	if logtype == "" {
		logtype = "text"
	}

	if logtype == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else if logtype == "text" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.WithField("logtype", logtype).Fatal("Given logtype was not valid, check LOG_TYPE configuration")
		os.Exit(1)
	}

	loglevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if len(loglevel) == 0 {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "warn" {
		log.SetLevel(log.WarnLevel)
	} else if loglevel == "error" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	return log.WithField("context", "shaker")
}

func (sh *Shaker) getShakerConfig (file string) []byte {
	conigByte, err := ioutil.ReadFile(file)
	if err != nil {
		sh.Log().Fatalf("Cant't read config file %s", file)
	}

	var config Config
	yaml.Unmarshal(conigByte, &config)
	byteConfig, _ := json.Marshal(config)


	return []byte(byteConfig)
}

func (sh *Shaker) getCronList (configByte []byte) {

	jobrunner.Start()
	var config Config
	yaml.Unmarshal(configByte, &config)

	sh.redisClient = sh.redisConnect(config.Redis.Host, config.Redis.Port, config.Redis.Password)

	for _, cronConfig := range config.Applications {
		sh.Log().Infoln("Prefix", cronConfig.Prefix)
		sh.Log().Infoln("Config", cronConfig.Config)
		configByte, err := ioutil.ReadFile(cronConfig.Config)
		if err != nil {
			sh.Log().Fatalf("Cant't read config file %s", cronConfig.Config)
		}
		var cronData CronData
		err = yaml.Unmarshal(configByte, &cronData)
		if err != nil {
			sh.Log().Fatal(err)
		}
		for _, data := range cronData {
			lockTimeout := 60
			if data.LockTimeout > 0 {
				lockTimeout = data.LockTimeout
			}
			sh.Log().Infof("Add job %s with lock timeout %d second", data.Name, lockTimeout)
			jobrunner.Schedule(data.Cron, RunJob{
				data.Name,
				string(cronConfig.Prefix + data.URI),
				sh.Log(),
				sh.redisClient,
				lockTimeout,
			})
		}

	}
}

func (sh *Shaker) Log() *log.Entry {
	return sh.log
}

func (sh *Shaker) Version() string {
	return sh.version
}

func (sh *Shaker) Stop() {
	sh.Log().Info("shutting things down")
	close(sh.stopCh)
}

func (sh *Shaker) Run() {
	for _, job := range sh.Jobs {
		go job.Run()
	}
}

func (sh *Shaker) redisConnect(host string, port string, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
		Password: password, // no password set
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		sh.Log().Fatalf("Can't connect redis: %s", err)
	}
	return client
}

func (e RunJob) Run() {
	locker := lock.New(e.redisClient, GetMD5Hash(e.URL), &lock.Options{
		LockTimeout: time.Second * 600,
		RetryCount: 0,
		RetryDelay: time.Microsecond * 100})

	// Try to obtain lock
	hasLock, err := locker.Lock()
	if err != nil {
		e.log.Panic(err.Error())
	} else if !hasLock {
		e.log.Infof("Job %s is already locked", e.Name)
		return
	}

	start := time.Now()
	resp, err := http.Get(e.URL)
	if err != nil {
		e.log.Error(err)
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Seconds()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.log.Errorf("Error: %s", err)
	}
	e.log = log.WithFields(log.Fields{
		"context": "shaker",
		"response": resp.StatusCode,
		"response_time": elapsed,
		"request": e.URL,
	})
	e.log.Info(string(body))
}

func (sh *Shaker) GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		sh.Log().Infof("Response code: %d Request URl: %s",c.Writer.Status(), c.Request.URL )
	}
}


func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}