package shaker

import (
	log "github.com/sirupsen/logrus"
)

func CreateRunnerContainer(config Config, log *log.Entry) map[string]JobRunner {
	var container = make(map[string]JobRunner)

	container["once"] = &SingleRunner{}
	container["reliable"] = &ReliableRunner{}

	for _, runner := range container {
		err := runner.Init(config, log)
		if err != nil {
			log.Errorf("Error: %s", err)
		}
	}

	return container
}
