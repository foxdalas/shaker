package shaker

import (
	"github.com/bsm/redis-lock"
	log "github.com/sirupsen/logrus"
	"net/http"
	"io/ioutil"
	"time"
)

func makeHTTP(e RunJob) {
	locker := lock.New(e.redisLock, GetMD5Hash(e.URL), &lock.Options{
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
			"description": e.Name,
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