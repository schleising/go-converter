package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/schleising/go-ffmpeg"
)

type Server struct {
	// An http.Server instance
	httpServer *http.Server

	// Channel to request progress information
	requestChannel chan struct{}

	// Channel to recieve progress information
	progressChannel chan go_ffmpeg.Progress
}

func NewServer() *Server {
	// Create a channel to request progress information
	requestChannel := make(chan struct{})

	// Create a channel to receive progress information
	progressChannel := make(chan go_ffmpeg.Progress)

	// Create a server instance
	server := Server{}

	// Create a handler function
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the progress information
		progress := server.getProgress()

		// Marshal the progress information into JSON
		progressBytes, err := json.Marshal(progress)

		// Check for errors
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the content type header to application/json
		w.Header().Set("Content-Type", "application/json")

		// Write a response
		w.Write(progressBytes)
	})

	// Create an http.Server instance
	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	// Set the server's fields
	server.httpServer = httpServer
	server.requestChannel = requestChannel
	server.progressChannel = progressChannel

	// Return the server
	return &server
}

func (s *Server) Start() {
	// Create a goroutine to listen for requests
	go func() {
		s.httpServer.ListenAndServe()
	}()
}

func (s *Server) Stop() error {
	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	// Defer cancelling the context
	defer cancel()

	// Shutdown the server
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) getProgress() go_ffmpeg.Progress {
	// Send a request for progress information
	s.requestChannel <- struct{}{}

	// Wait for progress information
	return <-s.progressChannel
}
