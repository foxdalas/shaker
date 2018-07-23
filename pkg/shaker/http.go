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
		"description": e.request.name,
		"context":     "shaker",
		"request":     e.request.url,
		"method":      "GET",
		"username":    e.request.username,
	})

	ok, err := e.lock.Lock()
	if err != nil {
		e.log.Errorf("Can't create lock %s", err)
		return
	} else if !ok {
		e.log.Debugf("Job %s is already locked", e.request.name)
		return
	}
	e.log.Debugf("Lock for job %s is created", e.request.name)

	start := time.Now()
	req, err := http.NewRequest("GET", e.request.url, nil)
	if err != nil {
		e.log.Error(err)
	}
	if e.request.username != "" || e.request.password != "" {
		req.SetBasicAuth(e.request.username, e.request.password)
	}
	cli := &http.Client{
		Timeout: e.request.timeout,
	}
	resp, err := cli.Do(req)
	if err != nil {
		e.log.Error(err)
		slackSendErrorMessage(e.clients.slackClient, e.request.name, err.Error(), e.request.url, 0)
		return
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Seconds()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		e.log.Error(err)
		slackSendErrorMessage(e.clients.slackClient, e.request.name, err.Error(), e.request.url, elapsed)
		return
	}
	e.log = log.WithFields(log.Fields{
		"context":       "shaker",
		"response_code": resp.StatusCode,
		"response_time": elapsed,
		"request":       e.request.url,
		"method":        "GET",
		"username":      e.request.username,
	})
	e.log.Info(string(body))

	checkResponseStatusCode(e, resp.StatusCode, elapsed)
	checkResponseBody(e, string(body), elapsed)
}

func checkResponseStatusCode(e RunJob, code int, elapsed float64) {
	message := fmt.Sprintf("Response code: %d", code)

	if code > 299 && code < 400 {
		slackSendWarningMessage(e.clients.slackClient, e.request.name, message, e.request.url, elapsed)
	}

	if code > 399 && code < 500 {
		slackSendWarningMessage(e.clients.slackClient, e.request.name, message, e.request.url, elapsed)
	}
	if code > 499 {
		slackSendErrorMessage(e.clients.slackClient, e.request.url, message, e.request.url, elapsed)
	}
}

func checkResponseBody(e RunJob, body string, elapsed float64) {
	if strings.Contains(body, "<script") {
		slackSendWarningMessage(e.clients.slackClient, e.request.name, "HTML in Response", e.request.url, elapsed)
	}
}
