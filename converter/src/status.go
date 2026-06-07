package main

import (
	"sync/atomic"

	"github.com/schleising/go-ffmpeg"
)

type Status struct {
	go_ffmpeg.Progress
	FilesRemaining int `json:"filesRemaining"`
}

type QueueTracker struct {
	remaining atomic.Int32
}

func (q *QueueTracker) Add() {
	q.remaining.Add(1)
}

func (q *QueueTracker) Done() {
	q.remaining.Add(-1)
}

func (q *QueueTracker) Remaining() int {
	return int(q.remaining.Load())
}

func newStatus(progress go_ffmpeg.Progress, queue *QueueTracker) Status {
	return Status{
		Progress:       progress,
		FilesRemaining: queue.Remaining(),
	}
}
