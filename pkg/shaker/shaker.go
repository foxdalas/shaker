package shaker

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"strings"
	"os"
	"github.com/foxdalas/shaker/pkg/shaker_const"
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
	"github.com/fsnotify/fsnotify"
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

	err := sh.params()
	if err != nil {
		sh.Log().Fatal(err)
	}
	sh.getCronList(sh.getShakerConfig(sh.configFile))
	go sh.watchJobs(sh.getShakerConfig(sh.configFile))
}

func (sh *Shaker) params() error {

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
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: log.FieldMap{
				log.FieldKeyMsg: "message",
				log.FieldKeyTime: "@timestamp",
			}})
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
	var config Config
	err := yaml.Unmarshal(configByte, &config)
	if err != nil {
		sh.log.Fatal(err)
		return
	}

	sh.redisClient = sh.redisConnect(config.Redis.Host, config.Redis.Port, config.Redis.Password)

	//Reading HTTP Jobs
	sh.log.Infof("Reading directory %s", config.Jobs.Http.Dir)
	files, err := ioutil.ReadDir(config.Jobs.Http.Dir)
	if err != nil {
		sh.log.Fatal(err)
		return
	}

	validate := sh.validaConfigs(config)
	if !validate {
		sh.Log().Error("Error in configration")
		return
	}

	for _, job := range jobrunner.Entries() {
		sh.Log().Infof("Cleanup job with id %d", job.ID)
		jobrunner.Remove(job.ID)
	}

	for _, file := range files {
		jobFile := config.Jobs.Http.Dir +"/"+file.Name()
		sh.Log().Infof("Reading file for HTTP jobs %s",jobFile)
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			sh.Log().Fatalf("Cant't read config file %s", jobFile)
		}
		var httpJobs HTTPJobs
		err = yaml.Unmarshal(configByte, &httpJobs)
		if err != nil {
			sh.Log().Fatal(err)
		}
		for _, data := range httpJobs.Jobs {
			lockTimeout := 30
			if data.LockTimeout > 0 {
				lockTimeout = data.LockTimeout
			}
			sh.Log().Infof("Add job %s with lock timeout %d second from file %s", data.Name, lockTimeout, jobFile)

			username := ""
			password := ""

			if len(data.Username) > 0 {
				username = data.Username
				if len(config.Users[username].Password) > 0 {
					password = config.Users[username].Password
				}
			}

			jobrunner.Schedule(data.Cron, RunJob{
				data.Name,
				string(httpJobs.URL + data.URI),
				"http",
				"get",
				username,
				password,
				sh.Log(),
				sh.redisClient,
				lockTimeout,
				jobFile,
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
		LockTimeout: time.Second * 300,
		RetryCount: 0,
		RetryDelay: time.Microsecond * 100})

	if locker.IsLocked() {
		e.log = log.WithFields(log.Fields{
			"context": "shaker",
			"request": e.URL,
		})
		e.log.Infof("Job %s is already locked", e.Name)
		return
	}

	start := time.Now()
	req, err := http.NewRequest("GET", e.URL, nil)
	if err != nil {
		e.log.Error(err)
	}
	if len(e.Username) > 0 || len(e.Password) >0  {
		req.SetBasicAuth(e.Username, e.Password)
	}
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		e.log = log.WithFields(log.Fields{
			"context":  "shaker",
			"error":    err,
			"request":  e.URL,
			"method":   "GET",
			"username": e.Username,
		})
		e.log.Error(err)
		return
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Seconds()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.log.Errorf("Error: %s", err)
		return
	}
	e.log = log.WithFields(log.Fields{
		"context": "shaker",
		"response_code": resp.StatusCode,
		"response_time": elapsed,
		"request": e.URL,
		"method": "GET",
		"username": e.Username,
	})
	e.log.Info(body)
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

func (sh *Shaker) watchJobs(configByte []byte) {
	var config Config
	err := yaml.Unmarshal(configByte, &config)
	if err != nil {
		sh.log.Fatal(err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				sh.Log().Infof("event:", event.String())
				sh.getCronList(configByte)
			case err := <-watcher.Errors:
				sh.Log().Errorf("error:", err)
			}
		}
	}()

	err = watcher.Add(config.Jobs.Http.Dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func (sh *Shaker) validaConfigs(config Config) bool {
	sh.log.Infof("Reading directory %s", config.Jobs.Http.Dir)
	files, err := ioutil.ReadDir(config.Jobs.Http.Dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		jobFile := config.Jobs.Http.Dir + "/" + file.Name()
		sh.Log().Infof("Reading file for HTTP jobs %s", jobFile)
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			sh.Log().Fatalf("Cant't read config file %s", jobFile)
			return false
		}
		var httpJobs HTTPJobs
		err = yaml.Unmarshal(configByte, &httpJobs)
		if err != nil {
			sh.Log().Error(err)
			return false
		}
	}
	return true
}

