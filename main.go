package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
	"github.com/segmentio/ksuid"
)

func init() {
	mkDirIfDoesntExist("data")
	files, err := ioutil.ReadDir("data")
	if err != nil {
		log.Fatalln("Couldnt list contents of data directory")
		return
	}
	for _, file := range files {
		id, err := ksuid.Parse(file.Name())
		if err != nil {
			log.Printf("Couln't parse id: %s", file.Name())
		}
		go deleteWhenRetentionRunsOut(id)
	}
}

//TODO: Middleware för att validera filnamnet
//TODO: Bygg stöd för path istället för filnamn. kunna lista directories
// TODO: Validate retention. min/max, type. konvertera till en duration
func main() {
	router := gin.Default()
	router.Use(location.Default())
	router.Use(isCurl())
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Static("/assets/", "./public/")
	router.LoadHTMLGlob("templates/*")

	upload := router.Group("/")
	{
		upload.GET("/", indexHandler)
		upload.POST("/", uploadHandler)
	}

	download := router.Group("/")
	{
		download.Use(validateID())
		download.GET("/g/:id/*path", listHandler)
		download.GET("/g/:id", listHandler)
		download.GET("/d/:id/*path", downloadHandler)
		download.GET("/z/:id", zipHandler)
		download.GET("/v/:id/*path", viewHandler)
	}

	router.GET("/ua", func(c *gin.Context) {
		ua := user_agent.New(c.Request.UserAgent())
		browserName, _ := ua.Browser()
		c.JSON(http.StatusOK, gin.H{
			"browser": browserName,
		})
	})
	router.Run()
}

func mkDirIfDoesntExist(target string) {
	if _, err := os.Stat(target); os.IsNotExist(err) {
		os.Mkdir(target, 0700)
	}
}

func deleteWhenRetentionRunsOut(id ksuid.KSUID) {
	durationLeft := durationLeft(id)
	// TODO: Handle error?

	if durationLeft.Seconds() > 0 {
		fmt.Printf("Sleeping for %s before deleting\n", durationLeft.String())
		time.Sleep(durationLeft)
	}
	fmt.Printf("Deleting %s\n", id)
	os.RemoveAll(fmt.Sprintf("data/%s", id))
}

func durationLeft(id ksuid.KSUID) time.Duration {
	var durationLeft time.Duration

	deletionTime := id.Time()
	now := time.Now()
	durationLeft = deletionTime.Sub(now)
	return durationLeft
}

func newUploadID(retention time.Duration) ksuid.KSUID {
	timeWithRetention := time.Now().Add(retention)
	ksuidID, err := ksuid.NewRandomWithTime(timeWithRetention)
	if err != nil {
		log.Panicf("Couldn't generate new id: %s", err.Error())
	}
	return ksuidID
}

func humanReadableFileSize(sizeInBytes int64) string {
	floatSize := float64(sizeInBytes)
	if sizeInBytes > 1E9 {
		return fmt.Sprintf("%.2f GB", roundPlaces(floatSize/1E9, 2))
	}
	if sizeInBytes > 1E6 {
		return fmt.Sprintf("%.2f MB", roundPlaces(floatSize/1E6, 2))
	}
	if sizeInBytes > 1E3 {
		return fmt.Sprintf("%.2f KB", roundPlaces(floatSize/1E3, 2))
	}
	return fmt.Sprintf("%d bytes", sizeInBytes)
}

func round(f float64) float64 {
	return math.Floor(f + .5)
}

func roundPlaces(f float64, places int) float64 {
	shift := math.Pow(10, float64(places))
	return round(f*shift) / shift
}
