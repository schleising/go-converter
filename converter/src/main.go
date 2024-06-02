package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/schleising/go-ffmpeg"
)

// Version of the converter
var version string = "0.1.2"

func main() {
	// Print the version
	fmt.Println("Converter Version: ", version)

	// Create a channel to receive the converter jobs
	converterJobChannel := make(chan Converter, 100)

	// Create a progress channel
	progressChannel := make(chan go_ffmpeg.Progress)

	// Create a wait group to wait for the goroutine to finish
	wg := sync.WaitGroup{}
	wg.Add(1)

	// Create a goroutine to convert files
	go func() {
		// Loop forever
		for job := range converterJobChannel {
			// Check if the context has been cancelled
			select {
			case <-job.ctx.Done():
				// Print a message to indicate that the conversion has been cancelled
				fmt.Println("Conversion cancelled:", job.inputFile)

				// Send an empty progress struct to indicate that the conversion is complete
				progressChannel <- go_ffmpeg.Progress{}
			default:
				// Print a message to indicate that the conversion has started
				fmt.Println("Converting file:", job.inputFile)

				// Convert the file
				err := job.convert()

				// Check for errors
				if err != nil {
					// Print an error message
					fmt.Println("Error converting file:", job.inputFile, err)
				} else {
					// Print a message to indicate that the conversion is complete
					fmt.Println("Conversion complete:", job.inputFile)

					// Call the cancel function
					job.cancelFunc()

					// Send an empty progress struct to indicate that the conversion is complete
					progressChannel <- go_ffmpeg.Progress{}
				}
			}
		}

		// Close the progress channel
		close(progressChannel)

		// Decrement the wait group
		wg.Done()
	}()

	// Create an empty map for Converter jobs
	jobs := make(map[string]*Converter)

	// Create a channel to listen for notifications
	notifyChannel := make(chan os.Signal, 1)

	// Notify the channel on interrupt or terminate
	signal.Notify(notifyChannel, syscall.SIGINT, syscall.SIGTERM)

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
		// Get the Downloads directory
		directory := "/Conversions"

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
		case <-notifyChannel:
			// Got a signal to close the server
			// Cancel all the jobs
			for _, job := range jobs {
				fmt.Println("Cancelling job:", job.inputFile)
				job.cancelFunc()
			}

			// Close the converter job channel
			close(converterJobChannel)

			// Set the closing variable to true
			closing = true
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

	// Wait for the goroutine to finish
	wg.Wait()

	// Stop the server
	err := server.Stop()

	// Check for errors
	if err != nil {
		fmt.Println("Error stopping server:", err)
	} else {
		fmt.Println("Application terminated successfully")
	}
}
