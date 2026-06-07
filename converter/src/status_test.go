package main

import (
	"testing"

	"github.com/schleising/go-ffmpeg"
)

func TestNewStatus(t *testing.T) {
	queue := &QueueTracker{}
	queue.Add()
	queue.Add()

	status := newStatus(go_ffmpeg.Progress{InputFile: "input.mkv"}, queue)

	if status.InputFile != "input.mkv" {
		t.Errorf("InputFile = %q, want input.mkv", status.InputFile)
	}
	if status.FilesRemaining != 2 {
		t.Errorf("FilesRemaining = %d, want 2", status.FilesRemaining)
	}
}

func TestQueueTrackerDone(t *testing.T) {
	queue := &QueueTracker{}
	queue.Add()
	queue.Done()

	if queue.Remaining() != 0 {
		t.Errorf("Remaining() = %d, want 0", queue.Remaining())
	}
}
