package main

import (
	"context"
	"fmt"

	"github.com/schleising/go-ffmpeg"
)

// Structure to hold a converter job instance
type Converter struct {
	// File to convert
	inputFile string

	// Channel to send progress information
	progressChannel chan go_ffmpeg.Progress

	// Context for the converter job
	ctx context.Context

	// Cancel function for the converter job
	cancelFunc context.CancelFunc
}

// Create a new converter instance
func NewConverter(inputFile string, progressChannel chan go_ffmpeg.Progress, ctx context.Context, cancelFunc context.CancelFunc) *Converter {
	// Create a new converter instance
	converter := Converter{
		inputFile:       inputFile,
		progressChannel: progressChannel,
		ctx:             ctx,
		cancelFunc:      cancelFunc,
	}

	// Return the converter instance
	return &converter
}

// Convert a file using ffmpeg
func (converter *Converter) convert() error {
	// Create a new Ffmpeg instance
	ffmpeg, err := go_ffmpeg.NewFfmpeg(
		converter.ctx,
		converter.inputFile,
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
				converter.progressChannel <- progress
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
