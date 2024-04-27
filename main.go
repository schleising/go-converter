package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func main() {
	// Create a channel to request a file to convert
	requestChannel := make(chan struct{})

	// Create a channel to receive the filename
	filenameChannel := make(chan string)

	// Create a goroutine to convert files
	go func() {
		// Loop forever
		for {
			// Request a file
			requestChannel <- struct{}{}

			// Wait for a filename
			filename := <-filenameChannel

			// Convert the file
			err := convert(filename)

			// Check for errors
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Conversion complete: ", filename)
			}
		}
	}()

	// Boolean to indicate that a file has been requested
	fileRequested := false

	// Create an empty set for files witha  boolean to indicate whether the file has been sent to the filename channel
	files := make(map[string]bool)

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
			if _, ok := files[newFile]; !ok {
				// Add the file to the set
				files[newFile] = false
			}
		}

		// Remove files that no longer exist
		for file := range files {
			// Check if the file exists
			if _, err := os.Stat(file); os.IsNotExist(err) {
				// Remove the file from the set
				delete(files, file)
			}
		}

		// Check whether there is a request for a file
		select {
		case <- requestChannel:
			// Set the fileRequested boolean to true
			fileRequested = true
		default:
			// Sleep for 100 milliseconds
			time.Sleep(100 * time.Millisecond)
		}

		// Check if a file has been requested
		if fileRequested {
			// If there is a request, send a filename that has not already been sent to the filename channel
			for file, sent := range files {
				if !sent {
					// Send the filename to the filename channel
					filenameChannel <- file

					// Mark the file as sent
					files[file] = true

					// Reset the fileRequested boolean
					fileRequested = false

					// Break out of the loop
					break
				}
			}
		}
	}
}
