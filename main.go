package main

/* Copyright 2017 Красимир Беров  */

/*
imgresize - A naive script for resizing images in a directory in depth

Currently imgresize can be invoked with parameter `-dir`
and it will traverse it in depth, search for files with
extension jpg or jpeg and create their resized copies
with max with 800 and max height 700 pixels.

TODO

* Add additional parameters for the required max with and height;
* Support PNG and GIF images;
* Do not process files if they are already processed.

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

var rejpg = regexp.MustCompile(`(?i)^(.+)\.jp(e)?g$`)

func main() {
	dir := flag.String("dir", "./", "root folder to search for images")
	flag.Parse()
	Println("Looking into folder:", *dir)
	FindFile(*dir, ProcessFile)
}

// ProcessFile resizes a particular file in its own goroutine
func ProcessFile(dir string, f os.FileInfo, wg sync.WaitGroup) {
	path := filepath.Join(dir, f.Name())
	if !rejpg.MatchString(path) {
		Println("Skipping ", path, "not a JPEG file")
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
	// Resize the image to max width 800 px and max height 700px
	// keeping the aspect ratio.
	img = resize.Resize(800, 0, img, resize.Lanczos3)
	img = resize.Resize(0, 700, img, resize.Lanczos3)
	// Create a copy and save it
	matches := rejpg.FindStringSubmatch(path)
	newFileName := matches[1] + "-800x700.jpg"
	fileOut, err := os.Create(newFileName)
	if err != nil {
		log.Println(err)
		return
	}
	defer fileOut.Close()
	defer wg.Done()
	Println("Creating", newFileName)
	//Write the new image to the newly created file
	jpeg.Encode(fileOut, img, nil)
}

// FindFiles finds files in `dir` and processes them using `wanted`
func FindFiles(dir string,
	wanted func(dir string, f os.FileInfo, wg sync.WaitGroup)) {
	// https://stackoverflow.com/questions/18207772/#18207832
	var wg = new(sync.WaitGroup)
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		wg.Add(1)
		go wanted(dir, f, wg)
		if f.IsDir() {
			FindFiles(filepath.Join(dir, f.Name()), wanted)
		}
	}
	wg.Wait()

}
