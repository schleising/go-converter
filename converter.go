package main

import (
	"context"
	"fmt"

	"github.com/schleising/go-ffmpeg"
)

func convert(inputFile string, progressChannel chan go_ffmpeg.Progress) error {
	// Create a context with a cancel function
	ctx, cancelFunc := context.WithCancel(context.Background())

	// Defer the cancel function
	defer cancelFunc()

	// Create a new Ffmpeg instance
	ffmpeg, err := go_ffmpeg.NewFfmpeg(
		ctx,
		inputFile,
		[]string{
			"-c:v", "libx264",
			"-c:a", "copy",
			"-c:s", "copy",
		},
	)

	// Check for errors
	if err != nil {
		return err
	}

	// Create a goroutine to listen for progress and errors
	go func() {
		for {
			select {
			// Listen for progress, requests, and errors
			case progress, ok := <-ffmpeg.Progress:
				if !ok {
					return
				}
				progressChannel <- progress
			case err, ok := <-ffmpeg.Error:
				if !ok {
					return
				}
				fmt.Println("Parsing Error:", err)
			}
		}
	}()

	// Start the ffmpeg process
	err = ffmpeg.Start()

	// Check for errors
	if err != nil {
		return err
	}

	// Wait for the ffmpeg process to finish
	<-ffmpeg.Done

	// Return nil
	return nil
}
