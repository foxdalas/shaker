package shaker
//@todo: Move to separate module

import (
	"io/ioutil"
	"encoding/json"
	"time"
)

type Job struct {
	Name     string        `json:"name"`
	Url      string        `json:"url"`
	Runner   string        `json:"runner"`
	Duration time.Duration `json:"duration"`
}

type Config struct {
	Environment string `json:"environment"`
	Role        string `json:"role"`
	Storage struct {
		Redis struct {
			Memory struct {
				Host string `json:"host"`
				Port int    `json:"port"`
			} `json:"memory"`
			Pubsub struct {
				Host string `json:"host"`
				Port int    `json:"port"`
			} `json:"pubsub"`
		} `json:"redis"`
	} `json:"storage"`
	Applications []struct {
		Prefix string `json:"prefix"`
		Config string `json:"config"`
	} `json:"applications"`
	Master struct {
		Socket string `json:"socket"` //@see: http://api.zeromq.org/4-1:zmq-connect#toc2
	} `json:"master"`
	IsMaster bool  `json:"isMaster"`
	Jobs     []Job `json:"jobs"`
}

func GetConfig(filePath string) (Config, error) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return Config{}, err
	}

	var config Config
	json.Unmarshal(file, &config)
	return config, nil
}
