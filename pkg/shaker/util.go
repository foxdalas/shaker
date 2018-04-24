package shaker

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

//MakeLog Create logging with fields
func MakeLog() *log.Entry {
	logtype := strings.ToLower(os.Getenv("LOG_TYPE"))
	if logtype == "" {
		logtype = "text"
	}

	if logtype == "json" {
		log.SetFormatter(&log.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: log.FieldMap{
				log.FieldKeyMsg:  "message",
				log.FieldKeyTime: "@timestamp",
			}})
	} else if logtype == "text" {
		log.SetFormatter(&log.TextFormatter{})
	} else {
		log.WithField("logtype", logtype).Fatal("Given logtype was not valid, check LOG_TYPE configuration")
		os.Exit(1)
	}

	loglevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if loglevel == "" {
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

//Log Entry
func (s *Shaker) Log() *log.Entry {
	return s.log
}

//Version return from Shaker struct
func (s *Shaker) Version() string {
	return s.version
}

//Run Job in routine
func (s *Shaker) Run() {
	for _, job := range s.jobs {
		go job.Run()
	}
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func urlFormater(url string, uri string) string {
	if url != "" && uri != "" {
		if url[len(url)-1:] != "/" && uri[:1] != "/" {
			return url + "/" + uri
		}
		return url + uri
	}
	return ""
}

func (s *Shaker) redisConnect(host string, port string, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: password, // no password set
		DB:       0,        // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		s.Log().Fatalf("Can't connect redis: %s", err)
	}
	return client
}
