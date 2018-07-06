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
		"description": e.request.name,
		"context":     "shaker",
		"channel":     e.request.channel,
		"method":      e.request.requestType,
		"type":        "redis",
	})

	if e.request.requestType == "publish" {
		client := e.clients.redisStorage
		err := client.Publish(e.request.channel, e.request.message).Err()
		if err != nil {
			e.log.Error(err)
			return
		}
		e.log.Info("ok")
	}
}
