package shaker

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"encoding/json"
	"github.com/bamzi/jobrunner"
)

func (s *Shaker) getConfig() {
	log := MakeLog()

	config, err := ioutil.ReadFile(s.configFile)
	if err != nil {
		log.Fatalf("Cant't read config file %s", s.configFile)
	}
	yaml.Unmarshal(config, &s.Config)
}

func (s *Shaker) isValidConfig() bool {
	if s.validateConfigs("http") && s.validateConfigs("redis") {
		return true
	}
	return false
}

func (s *Shaker) validateConfigs(jobType string) bool {
	var dir string
	var jobs Jobs

	switch jobType {
	case "http":
		dir = s.Config.Jobs.Http.Dir
	case "redis":
		dir = s.Config.Jobs.Redis.Dir
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
	var jobs Jobs

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
	if len(method) > 0 {
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


func (s *Shaker) loadJobs(jobs Jobs, jobFile string) {
	for _, data := range jobs.Jobs {
		lockTimeout := 0
		if data.Method != "publish" {
			lockTimeout = 30
			if data.LockTimeout > 0 {
				lockTimeout = data.LockTimeout
			}
		}
		s.Log().Infof("Add job %s with lock timeout %d second from file %s", data.Name, lockTimeout, jobFile)

		username := ""
		password := ""

		if len(data.Username) > 0 {
			username = data.Username
			if len(s.Config.Users[username].Password) > 0 {
				password = s.Config.Users[username].Password
			}
		}

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
			s.redisClient,
			lockTimeout,
			jobFile,
			s.RedisStorages[findRedisType(data.Method)],
		})
	}
}
