package shaker

import (
	log "github.com/sirupsen/logrus"
	"sync"
	"os"
	"github.com/foxdalas/shaker/pkg/shaker_const"
	"io/ioutil"
	"errors"
	"github.com/bamzi/jobrunner"
	"net/http"
	"time"
	"github.com/bsm/redis-lock"
	"github.com/fsnotify/fsnotify"
	"encoding/json"
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

func (s *Shaker) Init() {
	s.Log().Infof("Shaker %s starting", s.version)

	err := s.params()
	if err != nil {
		s.Log().Fatal(err)
	}

	s.getConfig()
	s.getCronList()

	go s.watchJobs()
}

func (s *Shaker) params() error {
	s.configFile = os.Getenv("CONFIG")
	if len(s.configFile) == 0 {
		return errors.New("Please provide the secret key via environment variable CONFIG")
	}
	return nil
}


func (s *Shaker) getCronList() {
	s.redisClient = s.redisConnect(s.Config.Redis.Host, s.Config.Redis.Port, s.Config.Redis.Password)

	//Reading HTTP Jobs
	s.log.Infof("Reading directory %s", s.Config.Jobs.Http.Dir)
	files, err := ioutil.ReadDir(s.Config.Jobs.Http.Dir)
	if err != nil {
		s.log.Fatal(err)
		return
	}

	if !s.validateConfigs() {
		s.Log().Error("Error in configration")
		return
	}

	for _, job := range jobrunner.Entries() {
		s.Log().Infof("Cleanup job with id %d", job.ID)
		jobrunner.Remove(job.ID)
	}

	for _, file := range files {
		jobFile := s.Config.Jobs.Http.Dir +"/"+file.Name()
		s.Log().Infof("Reading file for HTTP jobs %s",jobFile)
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			s.Log().Fatalf("Cant't read config file %s", jobFile)
		}

		var httpJobs HTTPJobs
		err = json.Unmarshal(configByte, &httpJobs)
		if err != nil {
			s.Log().Fatal(err)
		}

		for _, data := range httpJobs.Jobs {
			lockTimeout := 30
			if data.LockTimeout > 0 {
				lockTimeout = data.LockTimeout
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
				urlFormater(httpJobs.URL, data.URI),
				"http",
				"get",
				username,
				password,
				s.Log(),
				s.redisClient,
				lockTimeout,
				jobFile,
			})
		}
	}
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
	e.log.Info(string(body))
}

func (s *Shaker) watchJobs() {

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
				s.Log().Infof("event:", event.String())
				s.getCronList()
			case err := <-watcher.Errors:
				s.Log().Errorf("error:", err)
			}
		}
	}()

	err = watcher.Add(s.Config.Jobs.Http.Dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

