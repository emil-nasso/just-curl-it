package main

import (
	"fmt"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

// URLs
func downloadURL(id ksuid.KSUID, c *gin.Context, path string) string {
	return fmt.Sprintf("%s/d/%s/%s", baseURL(c), id, path)
}
func uploadURL(c *gin.Context) string {
	return baseURL(c)
}
func listURL(id ksuid.KSUID, c *gin.Context, path string) string {
	return fmt.Sprintf("%s/g/%s/%s", baseURL(c), id, path)
}
func viewURL(id ksuid.KSUID, c *gin.Context, path string) string {
	return fmt.Sprintf("%s/v/%s/%s", baseURL(c), id, path)
}
func zipURL(id ksuid.KSUID, c *gin.Context, path string) string {
	return fmt.Sprintf("%s/z/%s/%s", baseURL(c), id, path)
}
func baseURL(c *gin.Context) string {
	url := location.Get(c)
	return url.Scheme + "://" + url.Host
}

// Paths

func rootPath(id ksuid.KSUID) string {
	return fmt.Sprintf("data/%s/", id)
}
func filesPath(id ksuid.KSUID) string {
	return fmt.Sprintf("%s/files/", rootPath(id))
}
func zipPath(id ksuid.KSUID) string {
	return fmt.Sprintf("%s/%s.zip", rootPath(id), id)
}
func tarballPath(id ksuid.KSUID) string {
	return fmt.Sprintf("%s/%s.tar", rootPath(id), id)
}
func filePath(id ksuid.KSUID, filename string) string {
	return fmt.Sprintf("data/%s/files/%s", id, filename)
}
