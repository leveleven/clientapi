package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/zh-five/xdaemon"
)

type Network struct {
	Address    string   `form:"address" json:"address" url:"address" xml:"address" binding:"required"`
	Netmask    string   `form:"netmask" json:"netmask" url:"netmask" xml:"netmask" binding:"required"`
	Gateway    string   `form:"gateway" json:"gateway" url:"gateway" xml:"gateway" binding:"required"`
	DNS        []string `form:"dns" json:"dns" url:"dns" xml:"dns" binding:"required"`
	NeedReboot bool     `form:"need_reboot" json:"need_reboot"`
}

func main() {
	debug := os.Getenv("BOX_DEBUG")
	if debug != "on" {
		logFile := "clientapi.log"
		xdaemon.Background(logFile, true)
	}
	httpServer(debug)
}

func httpServer(debug string) {
	if debug != "on" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.GET("netcfg", func(c *gin.Context) {
		c.JSON(http.StatusOK, GetNetwork())
	})
	r.POST("netcfg", func(c *gin.Context) {
		var json Network
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			// return
		}
		if n := NetworkConfig(json.Address, json.Netmask, json.Gateway, json.DNS, json.NeedReboot); !n {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "500"})
			// return
		}
		if !json.NeedReboot {
			c.JSON(http.StatusOK, gin.H{
				"info": "restart machine to apply new configure."})
		}
	})
	r.GET("metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, metrics())
	})
	r.GET("ns", func(c *gin.Context) {
		c.JSON(http.StatusOK, getns())
	})
	r.Run(":8753")
}
