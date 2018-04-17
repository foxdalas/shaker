package shaker

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"encoding/json"
	)

func (s *Shaker) getConfig() {
	log := MakeLog()

	config, err := ioutil.ReadFile(s.configFile)
	if err != nil {
		log.Fatalf("Cant't read config file %s", s.configFile)
	}
	yaml.Unmarshal(config, &s.Config)
}

func (s *Shaker) validateConfigs() bool {
	s.log.Infof("Reading directory %s", s.Config.Jobs.Http.Dir)
	files, err := ioutil.ReadDir(s.Config.Jobs.Http.Dir)
	if err != nil {
		return false
	}

	for _, file := range files {
		jobFile := s.Config.Jobs.Http.Dir + "/" + file.Name()
		s.Log().Infof("Reading file for HTTP jobs %s", jobFile)
		configByte, err := ioutil.ReadFile(jobFile)
		if err != nil {
			s.Log().Fatalf("Cant't read config file %s", jobFile)
			return false
		}

		var httpJobs HTTPJobs

		err = json.Unmarshal(configByte, &httpJobs)
		if err != nil {
			s.Log().Error(err)
			return false
		}
	}
	return true
}