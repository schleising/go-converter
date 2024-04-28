package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/schleising/go-ffmpeg"
)

func main() {
	// Create a channel to receive the converter jobs
	converterJobChannel := make(chan Converter, 100)

	// Create a progress channel
	progressChannel := make(chan go_ffmpeg.Progress)

	// Create a goroutine to convert files
	go func() {
		// Loop forever
		for job := range converterJobChannel {
			// Check if the context has been cancelled
			select {
			case <-job.ctx.Done():
				// Print a message to indicate that the conversion has been cancelled
				fmt.Println("Conversion cancelled: ", job.inputFile)

				// Send an empty progress struct to indicate that the conversion is complete
				progressChannel <- go_ffmpeg.Progress{}
			default:
				// Convert the file
				err := job.convert()

				// Check for errors
				if err != nil {
					// Print an error message
					fmt.Println("Error converting file: ", job.inputFile)
					fmt.Println(err)
				} else {
					// Print a message to indicate that the conversion is complete
					fmt.Println("Conversion complete: ", job.inputFile)

					// Call the cancel function
					job.cancelFunc()

					// Send an empty progress struct to indicate that the conversion is complete
					progressChannel <- go_ffmpeg.Progress{}
				}
			}
		}

		// Close the progress channel
		close(progressChannel)
	}()

	// Create an empty map for Converter jobs
	jobs := make(map[string]*Converter)

	// Create a channel to listen for notifications
	notifyChannel := make(chan os.Signal, 1)

	// Notify the channel on interrupt or terminate
	signal.Notify(notifyChannel, syscall.SIGINT, syscall.SIGTERM)

	// Create a goroutine to listen for notifications
	go func() {
		// Wait for a notification
		<-notifyChannel

		// Cancel all the jobs
		for _, job := range jobs {
			job.cancelFunc()
		}

		// Close the converter job channel
		close(converterJobChannel)
	}()

	// Create a new server
	server := NewServer()

	// Start the server
	server.Start()

	// Create a progress and ok variable
	var progress go_ffmpeg.Progress
	var ok bool
	closing := false

	// Poll the directory for new files
	for {
		// Check for new files
		// If a new file is found, send the filename to the filename channel
		// If no new files are found, sleep for 100 milliseconds
		homeFolder, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			return
		}

		// Get the Downloads directory
		directory := filepath.Join(homeFolder, "Downloads")

		// Get a list of files in the directory with the extensions .mp4, .mkv, or .avi
		newFiles, err := filepath.Glob(filepath.Join(directory, "*.*"))
		if err != nil {
			fmt.Println(err)
			return
		}

		// Check for new files
		for _, newFile := range newFiles {
			// Check if the file is a .mp4, .mkv, or .avi file
			if filepath.Ext(newFile) != ".mp4" && filepath.Ext(newFile) != ".mkv" && filepath.Ext(newFile) != ".avi" {
				continue
			}

			// Check if the file is already in the set
			if _, ok := jobs[newFile]; !ok {
				// Create a context with a cancel function
				ctx, cancel := context.WithCancel(context.Background())

				// Create a new Converter instance
				converter := NewConverter(newFile, progressChannel, ctx, cancel)

				// Add the Converter instance to the map
				jobs[newFile] = converter

				// Send the Converter instance to the Converter channel
				converterJobChannel <- *converter
			}
		}

		// Remove files that no longer exist
		for file := range jobs {
			// Check if the file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				// Cancel the context
				jobs[file].cancelFunc()

				// Remove the file from the map
				delete(jobs, file)
			}
		}

		// Check whether there is a request for a file
		select {
		// Listen for progress, requests, and errors
		case progress, ok = <-progressChannel:
			if !ok {
				closing = true
			}
		case <-server.requestChannel:
			// Send the progress information
			server.progressChannel <- progress
		default:
			// Sleep for 100 milliseconds
			time.Sleep(100 * time.Millisecond)
		}

		// Check if the server is closing
		if closing {
			// Check if there are any requests for progress information before closing
			select {
			case <-server.requestChannel:
				// Send an empty progress struct
				server.progressChannel <- go_ffmpeg.Progress{}
			default:
			}
			break
		}
	}

	// Stop the server
	err := server.Stop()

	// Check for errors
	if err != nil {
		fmt.Println("Error stopping server")
		fmt.Println(err)
	}
}
