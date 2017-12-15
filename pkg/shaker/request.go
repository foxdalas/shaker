package shaker

import (
	"net/http"
	"io/ioutil"
)

func Fetch(url string) error {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if body == nil{

	}

	return err
}
