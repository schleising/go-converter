package main

import (
	"fmt"

	"github.com/schleising/go-ffmpeg"
)

func main() {
	ffmpeg, err := go_ffmpeg.NewFfmpeg(
		"/Users/steve/Downloads/TestInput.mp4",
		"/Users/steve/Downloads/Converted/TestOutput.mp4",
		[]string{
			"-c:v", "libx264",
			"-c:a", "copy",
			"-c:s", "copy",
		},
	)

	if err != nil {
		panic(err)
	}

	go func() {
		for progress := range ffmpeg.Progress {
			fmt.Println(progress)
		}
	}()

	err = ffmpeg.Start()

	if err != nil {
		panic(err)
	}
}
