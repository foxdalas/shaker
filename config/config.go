package config

import (
	"io/ioutil"
	"encoding/json"
)

func ReadConfig(filePath string, config interface{}) error {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	json.Unmarshal(file, &config)
	return nil
}
