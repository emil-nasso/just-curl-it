package main

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"github.com/segmentio/ksuid"
)

func validateID() gin.HandlerFunc {
	return func(c *gin.Context) {
		idString := c.Param("id")
		ksuidID, err := ksuid.Parse(idString)
		if err != nil {
			c.String(http.StatusNotFound, "Invalid id")
			c.Abort()
		} else {
			c.Set("uploadID", ksuidID)
			c.Next()
		}
	}
}

func isCurl() gin.HandlerFunc {
	return func(c *gin.Context) {
		ua := user_agent.New(c.Request.UserAgent())
		browserName, _ := ua.Browser()
		browserName = strings.ToLower(browserName)
		c.Set("isCurl", browserName == "curl" || browserName == "wget")
		c.Next()
	}
}
