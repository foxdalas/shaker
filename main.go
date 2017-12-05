package main

import (
	"github.com/foxdalas/shaker/pkg/shaker"
	"github.com/bamzi/jobrunner"
	"github.com/gin-gonic/gin"
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
	s := shaker.New(config, Version())
	s.Init()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(s.GinLogger(), gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.Data(200, "text/plain", []byte("pong"))
	})
	r.Run("127.0.0.1:8080")
}

func JobJson(c *gin.Context) {
	// returns a map[string]interface{} that can be marshalled as JSON
	c.JSON(200, jobrunner.StatusJson())
}

func JobHtml(c *gin.Context) {
	// Returns the template data pre-parsed
	c.HTML(200, "", jobrunner.StatusPage())

}