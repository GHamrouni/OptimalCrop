/*
Copyright (c) 2012, Ghassen Hamrouni <ghamrouni.iptech@gmail.com>

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.
*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"fsnotify"
	"image"
	"image/draw"
	"image/png"
	"imaging/optimalResize"
	"log"
	"os"
	"path/filepath"
	"time"
)

import _ "image/jpeg"
import _ "image/gif"

// Resize and crop an image
func resizeImage(inputFile string, outputFile string, width int, height int) {

	nTrials := 4

	success := false
	var imagefile *os.File

	// Only three image formats are supported (png, jpg, gif)
	if filepath.Ext(inputFile) == ".png" ||
		filepath.Ext(inputFile) == ".jpg" ||
		filepath.Ext(inputFile) == ".gif" {

		for nTrials > 0 && !success {

			success = true
			file, err := os.Open(inputFile)

			// TODO: implement a saner way to handle
			// oprn errors
			if err != nil {
				fmt.Printf("Open error \n")
				log.Println(err) // Don't use log.Fatal to exit

				success = false

				// give the system time to sync write change
				// sleep for some time before retry
				time.Sleep(500 * time.Millisecond)
			}

			nTrials--
			imagefile = file
		}

		// After multiple trials the system is unable
		// to access the file.
		if !success {
			return
		}

		defer imagefile.Close()

		// Decode the image.
		m, _, err := image.Decode(imagefile)
		if err != nil {
			fmt.Printf("Decode error \n")
			log.Println(err) // Don't use log.Fatal to exit

			return
		}

		b := m.Bounds()

		// All images are converted to the NRGBA type
		rgbaImage := image.NewNRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(rgbaImage, rgbaImage.Bounds(), m, b.Min, draw.Src)

		// Perform an optimal resize with 4 iterations  
		m2 := optimalResize.OptimalResize(rgbaImage, width, height, 4)

		fo, err := os.Create(outputFile)

		if err != nil {
			panic(err)
		}

		defer fo.Close()
		w := bufio.NewWriter(fo)

		defer w.Flush()
		png.Encode(w, m2)
	}
}

func main() {

	// Read the CMD options

	inDir := flag.String("in", "", "input directory")    // input directory
	outDir := flag.String("out", "", "output directory") // output directory
	width := flag.Int("width", 128, "the new width")     // width
	height := flag.Int("height", 128, "the new height")  // height

	flag.Parse()

	if *inDir == "" || *outDir == "" {
		log.Fatal("usage: \n imageResizer -in inputDir -out outputDir -width 128 -height 128")
	}

	// Print the cmd options

	fmt.Printf("image resize daemon \n")

	fmt.Printf("Input:  %s \n", *inDir)
	fmt.Printf("Output: %s \n", *outDir)
	fmt.Printf("Width:  %d \n", *width)
	fmt.Printf("Height: %d \n", *height)

	// Create a channel to use it to handle the idle state
	c := make(chan int)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	// Process file events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:

				log.Println("event:", ev)

				if ev.IsCreate() {
					// Combine the directory path with the filename to get 
					// the full path of the image
					fullImagePath := ev.Name
					fullImageOutPath := filepath.Join(*outDir, filepath.Base(ev.Name))

					// Launch asychronously a resize operation
					go resizeImage(fullImagePath, fullImageOutPath, *width, *height)
				}
			case err := <-watcher.Error:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Watch(*inDir)

	if err != nil {
		log.Fatal(err)
	}

	<-c
}
