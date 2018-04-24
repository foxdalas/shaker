# shaker - one more http cron manager

## Application config file
```
---
slack:
  enabled: true
  channel: cron
  token: secret-token
jobs:
  http:
    dir: "dist/jobs"
  redis:
    storages:
      default:
        host: 127.0.0.1
        port: 6379
      pubsub:
        host: 127.0.0.1
        port: 6379
    dir: "dist/redis"
users:
  user1:
    user: user1
    password: secret
  user2:
    user: user2
    password: secret
```
