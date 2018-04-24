package shaker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func makeHTTP(e RunJob) {
	e.log = log.WithFields(log.Fields{
		"description": e.Name,
		"context":     "shaker",
		"request":     e.URL,
		"method":      "GET",
		"username":    e.Username,
	})

	ok, err := e.lock.Lock()
	if err != nil {
		e.log.Errorf("Can't create lock %s", err)
		return
	} else if !ok {
		e.log.Debugf("Job %s is already locked", e.Name)
		return
	}
	e.log.Debugf("Lock for job %s is created", e.Name)

	start := time.Now()
	req, err := http.NewRequest("GET", e.URL, nil)
	if err != nil {
		e.log.Error(err)
	}
	if e.Username != "" || e.Password != "" {
		req.SetBasicAuth(e.Username, e.Password)
	}
	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		e.log.Error(err)
		slackSendErrorMessage(e.slack, e.Name, err.Error(), e.URL, 0)
		return
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Seconds()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.log.Error(err)
		slackSendErrorMessage(e.slack, e.Name, err.Error(), e.URL, elapsed)
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

	checkResponseStatusCode(e, resp.StatusCode, elapsed)
	checkResponseBody(e, string(body), elapsed)
}

func checkResponseStatusCode(e RunJob, code int, elapsed float64) {
	message := fmt.Sprintf("Response code: %d", code)

	if code > 299 && code < 400 {
		slackSendWarningMessage(e.slack, e.Name, message, e.URL, elapsed)
	}

	if code > 399 && code < 500 {
		slackSendWarningMessage(e.slack, e.Name, message, e.URL, elapsed)
	}
	if code > 499 {
		slackSendErrorMessage(e.slack, e.Name, message, e.URL, elapsed)
	}
}

func checkResponseBody(e RunJob, body string, elapsed float64) {
	if strings.Contains(body, "<script") {
		slackSendWarningMessage(e.slack, e.Name, "HTML in Response", e.URL, elapsed)
	}
}
