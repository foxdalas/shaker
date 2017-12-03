package shaker

import (
	log "github.com/sirupsen/logrus"
)

func CreateRunnerContainer(config Config, log *log.Entry) map[string]JobRunner {
	var container = make(map[string]JobRunner)

	container["once"] = &Once{}

	for _, runner := range container {
		runner.Init(config, log)
	}

	return container
}
