package shaker

import (
	log "github.com/sirupsen/logrus"
	"strings"
	"os"
	"time"
	"encoding/hex"
	"crypto/md5"
)

func MakeLog() *log.Entry {
	logtype := strings.ToLower(os.Getenv("LOG_TYPE"))
	if logtype == "" {
		logtype = "text"
	}

	if logtype == "json" {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: log.FieldMap{
				log.FieldKeyMsg: "message",
				log.FieldKeyTime: "@timestamp",
			}})
	} else if logtype == "text" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.WithField("logtype", logtype).Fatal("Given logtype was not valid, check LOG_TYPE configuration")
		os.Exit(1)
	}

	loglevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if len(loglevel) == 0 {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if loglevel == "info" {
		log.SetLevel(log.InfoLevel)
	} else if loglevel == "warn" {
		log.SetLevel(log.WarnLevel)
	} else if loglevel == "error" {
		log.SetLevel(log.ErrorLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	return log.WithField("context", "shaker")
}

func (sh *Shaker) Log() *log.Entry {
	return sh.log
}

func (sh *Shaker) Version() string {
	return sh.version
}

func (sh *Shaker) Run() {
	for _, job := range sh.Jobs {
		go job.Run()
	}
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}


func urlFormater(url string, uri string) string {
	if url[len(url)-1:] != "/" && uri[:1] != "/" {
		return url + "/" + uri
	} else {
		return url + uri
	}
}