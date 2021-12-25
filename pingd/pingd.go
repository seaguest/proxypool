package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seaguest/log"
	"github.com/seaguest/proxypool"
)

func main() {
	r := gin.New()

	gin.SetMode(gin.DebugMode)

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		type param struct {
			Ip string `form:"ip"`
		}
		var p param
		var err error
		if err = c.Bind(&p); err != nil {
			log.Error(err)
			return
		}

		remoteAddr := c.Request.Header.Get("REMOTE_ADDR")
		httpVia := c.Request.Header.Get("HTTP_VIA")
		forwardFor := c.Request.Header.Get("HTTP_X_FORWARDED_FOR")

		var anonymity int
		if forwardFor == "" && httpVia == "" && (remoteAddr == p.Ip || remoteAddr == "") {
			anonymity = proxypool.AnonymityHigh
		} else if forwardFor == p.Ip && httpVia == p.Ip && remoteAddr == p.Ip {
			anonymity = proxypool.AnonymityAnonymous
		} else {
			anonymity = proxypool.AnonymityTransparent
		}

		log.Error("----------", remoteAddr, httpVia, forwardFor)
		c.JSON(http.StatusOK, gin.H{"err_code": 0, "anonymity": anonymity})
	})

	r.Run(":9001")
}
