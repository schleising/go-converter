package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/schleising/go-ffmpeg"
)

func main() {
	// Create a context with a cancel function
	ctx, cancelFunc := context.WithCancel(context.Background())

	// Defer the cancel function
	defer cancelFunc()

	// Create a signal channel
	signalChannel := make(chan os.Signal, 1)

	// Notify the signal channel of any interrupt signals
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	// Create a goroutine to listen for signals
	go func() {
		// Wait for a signal
		<-signalChannel

		// Cancel the context
		cancelFunc()
	}()

	// Create a new Ffmpeg instance
	ffmpeg, err := go_ffmpeg.NewFfmpeg(
		ctx,
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
		for {
			select {
			case progress, ok := <-ffmpeg.Progress:
				if !ok {
					fmt.Println("Progress channel closed")
					return
				}
				fmt.Println(progress)
			case err, ok := <-ffmpeg.Error:
				if !ok {
					fmt.Println("Error channel closed")
					return
				}
				fmt.Println(err)
			}
		}
	}()

	err = ffmpeg.Start()

	if err != nil {
		fmt.Println(err)
	}
}
