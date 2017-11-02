package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mholt/archiver"
	"github.com/segmentio/ksuid"
)

func zipHandler(c *gin.Context) {
	// TODO: should only create zip if it doesnt exist.
	// TODO: should also support tarballs, tar.gzip?
	id := c.MustGet("uploadID").(ksuid.KSUID)
	filesDirectoryPath := filesPath(id)
	files, err := ioutil.ReadDir(filesDirectoryPath)
	if err != nil {
		log.Fatal(err)
	}

	filePaths := make([]string, 0)
	for _, file := range files {
		filePaths = append(filePaths, filesDirectoryPath+file.Name())
	}
	archivePath := zipPath(id)
	archiver.Zip.Make(archivePath, filePaths)
	c.File(archivePath)
}

func downloadHandler(c *gin.Context) {
	id := c.MustGet("uploadID").(ksuid.KSUID)
	path := c.Param("path")
	filename := filepath.Base(path)
	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filename)
	c.File(filePath(id, path))
}

func viewHandler(c *gin.Context) {
	id := c.MustGet("uploadID").(ksuid.KSUID)
	filename := c.Param("path")
	fmt.Println("Viewing fo sho")
	fmt.Println(filePath(id, filename))
	c.File(filePath(id, filename))
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"uploadUrl": uploadURL(c),
	})
}

func listHandler(c *gin.Context) {
	id := c.MustGet("uploadID").(ksuid.KSUID)
	path := c.Param("path")

	fullPath := filesPath(id) + path
	files, err := ioutil.ReadDir(fullPath)
	if err != nil {
		c.String(http.StatusNotFound, "404 Not Found")
		return
	}

	if c.GetBool("isCurl") {
		// TODO: Test and fix
		// if len(files) == 1 {
		// 	c.Redirect(http.StatusFound, downloadURL(id, c, files[0].Name()))
		// 	return
		// }
		// responseString := ""
		// for _, file := range files {
		// 	responseString = fmt.Sprintf("%s\n%s", responseString, downloadURL(id, c, file.Name()))
		// }
		// c.String(http.StatusOK, "Multiple files available for download:\n%s\n", responseString)
	} else {
		filesInfo := make([]gin.H, 0)
		var view, download, currentFilePath string
		for _, file := range files {
			currentFilePath = strings.Trim(path+"/"+file.Name(), "/")
			if file.IsDir() {
				view = listURL(id, c, currentFilePath)
				download = zipURL(id, c, currentFilePath)
			} else {
				view = viewURL(id, c, currentFilePath)
				download = downloadURL(id, c, currentFilePath)
			}
			filesInfo = append(filesInfo, gin.H{
				"name":        file.Name(),
				"isDir":       file.IsDir(),
				"size":        humanReadableFileSize(file.Size()),
				"viewURL":     view,
				"downloadURL": download,
			})
		}

		c.HTML(http.StatusOK, "get.tmpl", gin.H{
			"id":           id,
			"files":        filesInfo,
			"timeLeft":     durationLeft(id).String(),
			"parentURL":    listURL(id, c, strings.Trim(filepath.Dir(path), "/")),
			"hasParentDir": path != "" && path != "/",
		})
	}
}

// TODO: files are required. Don't accept posts without files
// TODO: Visa olika urler / kommandon för curl baserat på om man laddade upp
//       en eller flera filer
func uploadHandler(c *gin.Context) {
	var unit string
	if unit = c.PostForm("unit"); unit == "" {
		unit = "h"
	}

	var retention int
	var err error
	if retention, err = strconv.Atoi(c.PostForm("retention")); err != nil {
		retention = 24
	}

	if unit == "d" {
		unit = "h"
		retention *= 24
	}
	duration, err := time.ParseDuration(fmt.Sprintf("%d%s", retention, unit))
	if err != nil {
		c.String(http.StatusBadRequest, "Couldn't parse duration")
		return
	}

	id := newUploadID(duration)
	if err != nil {
		c.String(http.StatusBadRequest, "Couldnt generate id")
		return
	}
	// Multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}
	files := form.File["files"]

	mkDirIfDoesntExist(rootPath(id))
	mkDirIfDoesntExist(filesPath(id))

	if len(files) == 1 && files[0].Filename == "files.tar" {
		// TODO: Should also support zip/tar.gz?
		// TODO: do something to ignore "._" prefixed files in osx?
		tarBall := files[0]
		if err := c.SaveUploadedFile(tarBall, tarballPath(id)); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
			return
		}
		archiver.Tar.Open(tarballPath(id), filesPath(id))
	} else {
		for _, file := range files {
			if err := c.SaveUploadedFile(file, filePath(id, file.Filename)); err != nil {
				c.String(http.StatusBadRequest, fmt.Sprintf("upload file err: %s", err.Error()))
				return
			}
		}
	}

	url := listURL(id, c, "")
	if c.GetBool("isCurl") {
		c.String(http.StatusOK, "Download available at:\n %s\n", url)
	} else {
		c.Redirect(http.StatusFound, url)
	}

	go deleteWhenRetentionRunsOut(id)

}
