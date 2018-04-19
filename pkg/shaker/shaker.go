package shaker

import (
	"errors"
	"github.com/foxdalas/shaker/pkg/shaker_const"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
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
	s.createRedisConnections()
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

	//Checking configuration
	if !s.isValidConfig() {
		s.Log().Error("Error in configration")
		return
	}

	s.cleanupJobs()

	s.readConfigDirectory(s.Config.Jobs.Http.Dir, "http")
	s.readConfigDirectory(s.Config.Jobs.Redis.Dir, "redis")
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
				s.Log().Errorf("error: %s", err)
			}
		}
	}()

	err = watcher.Add(s.Config.Jobs.Http.Dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
