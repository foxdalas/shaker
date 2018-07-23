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

* slack - slack configuration
* jobs - http or redis pubsub "publish" jobs
* redis.storages.default - redis for destributed locks
* redis.storages.pubsub - redis for publish to pubsub jobs
* user - user/password for http jobs

## Job configuration

```
{
    "url": "http:/localhost",
    "jobs": [
        {
            "name": "Every minute in 0 second",
            "cron": "0 * * * *",
            "uri": "api/myMethod1"
         },
         {
             "name": "Every minute in 0 second 2 with 10 second timeout",
             "cron": "0 */1 * * * *",
             "uri": "api/myMethod2",
             "timeout": 10
         },
         {
             "name": "Every 4 minutes in 0 second",
             "cron": "0 */4 * * * *",
             "uri": "api/myMethod3"
         },
         {
             "name": "Every day in 1 hour with user user1",
             "cron": "0 0 1 * * *",
             "uri": "api/myMethod4",
             "username": "user1"
         }
    ]
}
```
