package shaker

import (
	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
)

func (s *Shaker) createRedisConnections() {
	connections := make(map[string]*redis.Client)

	for name, info := range s.config.Jobs.Redis.Storages {
		connections[name] = redis.NewClient(&redis.Options{
			Addr: info.Host + ":" + info.Port,
		})
	}

	for name, client := range connections {
		s.Log().Infof("Connection to redis %s", name)
		_, err := client.Ping().Result()
		if err != nil {
			s.Log().Errorf("%s: %s", name, err)
		}
	}

	s.connectors.redisStorages = connections
}

func makeRedis(e RunJob) {
	e.log = log.WithFields(log.Fields{
		"description": e.Name,
		"context":     "shaker",
		"channel":     e.Channel,
		"method":      e.Method,
		"type":        "redis",
	})

	if e.Method == "publish" {
		client := e.redisStorage
		err := client.Publish(e.Channel, e.Message).Err()
		if err != nil {
			e.log.Error(err)
			return
		}
		e.log.Info("ok")
	}
}