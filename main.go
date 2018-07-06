package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/aksentyev/shaker/pkg/shaker"
	"github.com/bamzi/jobrunner"
	"github.com/sirupsen/logrus"
)

var AppVersion = ""
var AppGitCommit = ""
var AppGitState = ""
var stopCh chan struct{}

func main() {
	jobrunner.Start()

	s := shaker.New(Version())
	s.Init()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	signal := <-c
	logger := logrus.WithField("signal", signal.String())
	logger.Debug("received signal")
	Stop()
}

//Stop application
func Stop() {
	logrus.Info("shutting things down")
	stopCh := make(chan struct{})
	close(stopCh)
}

//Version helper
func Version() string {
	version := AppVersion
	if AppGitCommit != "" {
		version += "-"
		version += AppGitCommit[0:8]
	}
	if AppGitState != "" && AppGitState != "clean" {
		version += "-"
		version += AppGitState
	}
	return version
}
