package main

/* Copyright 2017 Красимир Беров  */

/*
Command imgresize - A naive program for resizing images in a tree of directories

Each file is processed in its own goroutine, which means imgresize should be very fast.
Currently supported file formats are JPEG, GIF and PNG.

Usage of imgresize:
  -dir string
    	root folder to search for images (default "./")
  -maxheight uint
    	maximal height of the resized image (default 800)
  -maxwidth uint
    	maximal width of the resized image (default 800)

This program depends on github.com/nfnt/resize

*/

import (
	"flag"
	. "fmt"
	"github.com/nfnt/resize"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
)

var (
	maxwidth  *uint64
	maxheight *uint64
)
var extension = `(jp(e)?g|png|gif)$`
var rejpg = regexp.MustCompile(`(?i)^(.+)?\.` + extension)
var reresized = regexp.MustCompile(`(?i)\d+x\d+\.` + extension)

func main() {
	dir := flag.String("dir", "./", "root folder to search for images")
	maxwidth = flag.Uint64("maxwidth", 800, "maximal width of the resized image")
	maxheight = flag.Uint64("maxheight", 800, "maximal height of the resized image")
	flag.Parse()
	Println("Looking into folder:", *dir)
	FindFiles(*dir, ProcessFile)
}

// ProcessFile resizes a particular file in its own goroutine
func ProcessFile(dir string, f os.FileInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	path := filepath.Join(dir, f.Name())
	if !rejpg.MatchString(path) {
		Println("Skipping", path, "not a JPEG, PNG nor GIF file")
		return
	}
	if reresized.MatchString(path) {
		Println("Skipping", path, "This is a result of resizing.")
		return
	}
	matches := rejpg.FindStringSubmatch(path)
	newFileName := matches[1] + Sprintf(`-%dx%d.`+matches[2], *maxwidth, *maxheight)
	if _, err := os.Stat(newFileName); err == nil {
		Println("Skipping", path, "It is already resized as", newFileName)
		return
	}

	Println("resizing", path)
	fileIn, err := os.Open(path)
	if err != nil {
		log.Println(err)
		return
	}
	defer fileIn.Close()

	var img image.Image
	if ok, _ := regexp.MatchString("(?i)jp(e)?g$", path); ok {
		Println("decoding", path)
		img, err = jpeg.Decode(fileIn)
	} else if ok, _ := regexp.MatchString("(?i)png$", path); ok {
		Println("decoding", path)
		img, err = png.Decode(fileIn)
	} else if ok, _ := regexp.MatchString("(?i)gif$", path); ok {
		Println("decoding", path)
		img, err = gif.Decode(fileIn)
	}
	if err != nil {
		log.Println(err)
		return
	}
	// Resize the image to max width pixels and max height pixels keeping the aspect ratio.
	img = resize.Thumbnail(uint(*maxwidth), uint(*maxheight), img, resize.Lanczos3)
	// Create a copy and save it
	fileOut, err := os.Create(newFileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer fileOut.Close()
	Println("Creating", newFileName)
	//Write the new image to the newly created file
	if ok, _ := regexp.MatchString("(?i)jp(e)?g$", newFileName); ok {
		err = jpeg.Encode(fileOut, img, nil)
	} else if ok, _ := regexp.MatchString("(?i)png$", newFileName); ok {
		err = png.Encode(fileOut, img)
	} else if ok, _ := regexp.MatchString("(?i)gif$", newFileName); ok {
		err = gif.Encode(fileOut, img, nil)
	}
	if err != nil {
		log.Println(err)
	}
}

// FindFiles finds files in `dir` and processes them using `wanted`
func FindFiles(dir string, wanted func(dir string, f os.FileInfo, wg *sync.WaitGroup)) {
	// https://stackoverflow.com/questions/18207772/#18207832
	var wg = new(sync.WaitGroup)
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		log.Fatal(err)
	}
	var folders []string
	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, filepath.Join(dir, file.Name()))
			continue
		}
		wg.Add(1)
		go wanted(dir, file, wg)
	}
	wg.Wait()
	//process folders in this folder
	for _, f := range folders {
		FindFiles(f, wanted)
	}
}
