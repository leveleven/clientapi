package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Network struct {
	Address    string   `form:"address" json:"address" url:"address" xml:"address" binding:"required"`
	Netmask    string   `form:"netmask" json:"netmask" url:"netmask" xml:"netmask" binding:"required"`
	Gateway    string   `form:"gateway" json:"gateway" url:"gateway" xml:"gateway" binding:"required"`
	DNS        []string `form:"dns" json:"dns" url:"dns" xml:"dns" binding:"required"`
	NeedReboot bool     `form:"need_reboot" json:"need_reboot"`
}

func main() {
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
	r.Run(":8753")
}
