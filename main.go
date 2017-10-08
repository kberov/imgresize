package main

/* Copyright 2017 Красимир Беров  */

/*
Command imgresize - A naive script for resizing images in a directory in depth

Currently supports only `jp(e)?g` files.

Usage of imgresize:
  -dir string
    	root folder to search for images (default "./")
  -maxheight uint
    	maximal height of the resized image (default 800)
  -maxwidth uint
    	maximal width of the resized image (default 800)

TODO:

* Support PNG and GIF images;

*/

import (
	"flag"
	. "fmt"
	"github.com/nfnt/resize"
	"image/jpeg"
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
var rejpg = regexp.MustCompile(`(?i)^(.+)\.jp(e)?g$`)
var reresized = regexp.MustCompile(`(?i)\d+x\d+\.(jp(e)?g|png|gif)$`)

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
	matches := rejpg.FindStringSubmatch(path)
	newFileName := matches[1] + Sprintf(`-%dx%d.jpg`, *maxwidth, *maxheight)
	if !rejpg.MatchString(path) {
		Println("Skipping ", path, "not a JPEG, PNG nor GIF file")
		return
	}
	if reresized.MatchString(path) {
		Println("Skipping ", path, "This is a result of resizement.")
		return
	}
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
	img, err := jpeg.Decode(fileIn)
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
	jpeg.Encode(fileOut, img, nil)
}

// FindFiles finds files in `dir` and processes them using `wanted`
func FindFiles(dir string,
	wanted func(dir string, f os.FileInfo, wg *sync.WaitGroup)) {
	// https://stackoverflow.com/questions/18207772/#18207832
	var wg = new(sync.WaitGroup)
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		log.Fatal(err)
	}
	var folders []string
	added := 0
	for _, f := range files {
		if f.IsDir() {
			folders = append(folders, filepath.Join(dir, f.Name()))
			continue
		}
		wg.Add(1)
		go wanted(dir, f, wg)
		added++
		if added == 10 { //wait for ten files to be processed, then continue.
			wg.Wait()
			added = 0
		}
	}
	if added > 0 { //wait for the remaining less than ten files to be processed
		wg.Wait()
	}

	//process folders in this folder
	for _, f := range folders {
		FindFiles(f, wanted)
	}
}
