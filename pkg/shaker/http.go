package shaker

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

func makeHTTP(e RunJob) {
	if e.lock.IsLocked() {
		e.log = log.WithFields(log.Fields{
			"context": "shaker",
			"request": e.URL,
		})
		e.log.Infof("Job %s is already locked", e.Name)
		return
	}
	ok, err := e.lock.Lock()
	if err != nil {
		e.log = log.WithFields(log.Fields{
			"context": "shaker",
			"request": e.URL,
		})
		e.log.Errorf("Can't create lock %s", err)
		return
	} else if !ok {
		e.log = log.WithFields(log.Fields{
			"context": "shaker",
			"request": e.URL,
		})
		e.log.Errorf("Can't create lock for job %s", e.Name)
		return
	}

	e.log.Debugf("Lock for job %s is created", e.Name)
	
	start := time.Now()
	req, err := http.NewRequest("GET", e.URL, nil)
	if err != nil {
		e.log.Error(err)
	}
	if len(e.Username) > 0 || len(e.Password) > 0 {
		req.SetBasicAuth(e.Username, e.Password)
	}
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		e.log = log.WithFields(log.Fields{
			"description": e.Name,
			"context":     "shaker",
			"error":       err,
			"request":     e.URL,
			"method":      "GET",
			"username":    e.Username,
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
		"context":       "shaker",
		"response_code": resp.StatusCode,
		"response_time": elapsed,
		"request":       e.URL,
		"method":        "GET",
		"username":      e.Username,
	})
	e.log.Info(string(body))
}
