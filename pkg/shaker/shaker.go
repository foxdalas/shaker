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

//New create struct for shaker
func New(version string) *Shaker {
	return &Shaker{
		version:   version,
		log:       MakeLog(),
		stopCh:    make(chan struct{}),
		waitGroup: sync.WaitGroup{},
	}
}

//Init Shaker main functional
func (s *Shaker) Init() {
	s.Log().Infof("Shaker %s starting", s.version)

	err := s.params()
	if err != nil {
		s.Log().Fatal(err)
	}

	if s.isSlackEnabled() {
		s.createSlackConnection()
		slackSendInfoMessage(s.connectors.slackConfig, "Shaker service", "Started", "", 0)
	}

	s.createRedisConnections()
	s.getCronList()
	if s.config.Watch.Dir != "" {
		go s.watchJobs()
	}
}

func (s *Shaker) params() error {
	if os.Getenv("CONFIG") == "" {
		return errors.New("Please provide the secret key via environment variable CONFIG")
	}
	s.getConfig(os.Getenv("CONFIG"))

	return nil
}

func (s *Shaker) getCronList() {
	//Checking configuration
	if !s.isValidConfig() {
		s.Log().Error("Error in configration")
		return
	}
	//Cleanup jobs is configuration is valid
	s.cleanupJobs()

	//Add jobs from configuration files
	s.readConfigDirectory(s.config.Jobs.HTTP.Dir, "http")
	s.readConfigDirectory(s.config.Jobs.Redis.Dir, "redis")
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
				if event.Op&fsnotify.Create == fsnotify.Create {
					s.getCronList()
					slackSendInfoMessage(s.connectors.slackConfig, "Configuration", "Apply", "", 0)
				}
			case err := <-watcher.Errors:
				s.Log().Errorf("error: %s", err)
			}
		}
	}()

	err = watcher.Add(s.config.Watch.Dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
