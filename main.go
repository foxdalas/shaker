package main

import (
	"shaker/pkg/shaker"
	"github.com/gin-gonic/gin"
	"shaker/config"
)

var AppVersion = "unknown"
var AppGitCommit = ""
var AppGitState = ""

func Version() string {
	version := AppVersion
	if len(AppGitCommit) > 0 {
		version += "-"
		version += AppGitCommit[0:8]
	}
	if len(AppGitState) > 0 && AppGitState != "clean" {
		version += "-"
		version += AppGitState
	}
	return version
}

func main() {
	var conf shaker.Config
	config.ReadConfig("./config.json", &conf)
	s := shaker.New(conf, Version())
	s.Start()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	//r.Use(s.GinLogger(), gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("pong"))
	})
	r.Run("127.0.0.1:8080")
}
