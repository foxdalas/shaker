package shaker

import (
	"encoding/json"
	"github.com/bamzi/jobrunner"
	"github.com/bsm/redis-lock"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

func (s *Shaker) getConfig(configFile string) {
	log := MakeLog()
	s.Log().Infof("Reading configuration from %s", configFile)
	config, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Cant't read config file %s", configFile)
	}
	err = yaml.Unmarshal(config, &s.config)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Shaker) isValidConfig() bool {
	if s.validateConfigs("http") && s.validateConfigs("redis") {
		return true
	}
	return false
}

func (s *Shaker) validateConfigs(jobType string) bool {
	var dir string
	var jobs jobs

	switch jobType {
	case "http":
		dir = s.config.Jobs.HTTP.Dir
	case "redis":
		dir = s.config.Jobs.Redis.Dir
	}

	s.log.Infof("Reading directory %s", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		jobFile := dir + "/" + file.Name()
		s.Log().Infof("Reading file for %s jobs %s", jobType, jobFile)
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			s.Log().Fatalf("Cant't read config file %s", jobFile)
			return false
		}

		err = json.Unmarshal(configByte, &jobs)
		if err != nil {
			s.Log().Error(err)
			return false
		}
	}

	return true
}

func (s *Shaker) readConfigDirectory(dir string, jobType string) {
	var jobs jobs

	s.log.Infof("Reading directory %s", dir)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		s.log.Fatal(err)
		return
	}

	for _, file := range files {
		jobFile := dir + "/" + file.Name()
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			s.Log().Fatalf("Cant't read config file %s", jobFile)
		}

		err = json.Unmarshal(configByte, &jobs)
		if err != nil {
			s.Log().Error(err)
		}
		s.loadJobs(jobs, jobFile)
	}

}

func findType(method string) string {
	switch method {
	case "get":
		return "http"
	case "post":
		return "http"
	case "publish":
		return "redis"
	default:
		return "http"
	}
}

func findMethod(method string) string {
	if method != "" {
		return method
	}
	return "get"
}

func findRedisType(method string) string {
	if findType(method) == "redis" {
		switch findMethod(method) {
		case "publish":
			return "pubsub"
		}
	}
	return "default"
}

func (s *Shaker) loadJobs(jobs jobs, jobFile string) {
	for _, data := range jobs.Jobs {
		lockTimeout := 0
		if data.Method != "publish" {
			lockTimeout = 30
			if data.LockTimeout > 0 {
				lockTimeout = data.LockTimeout
			}
		}
		s.Log().Infof("Add job %s with lock timeout %d second from file %s", data.Name, lockTimeout, jobFile)

		var username string
		var password string

		if data.Username != "" {
			username = data.Username
			if s.config.Users[username].Password != "" {
				password = s.config.Users[username].Password
			}
		}

		locker := lock.New(s.connectors.redisStorages["default"], getMD5Hash(urlFormater(jobs.URL, data.URI)), &lock.Options{
			LockTimeout: time.Duration(lockTimeout) * time.Second,
			RetryCount:  0,
			RetryDelay:  time.Microsecond * 100})

		jobrunner.Schedule(data.Cron, RunJob{
			data.Name,
			urlFormater(jobs.URL, data.URI),
			jobs.Redis,
			findType(data.Method),
			findMethod(data.Method),
			username,
			password,
			data.Channel,
			data.Message,
			s.Log(),
			locker,
			jobFile,
			s.connectors.redisStorages[findRedisType(data.Method)],
			s.connectors.slackConfig,
		})
	}
}