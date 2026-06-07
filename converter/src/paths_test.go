package main

import (
	"path/filepath"
	"testing"
)

func TestOutputPathForInput(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/Conversions/movie.mkv",
			expected: "/Conversions/Converted/movie.mp4",
		},
		{
			input:    "/Conversions/MyShow/S01/ep01.mkv",
			expected: "/Conversions/Converted/MyShow/S01/ep01.mp4",
		},
	}

	for _, tt := range tests {
		output, err := outputPathForInput(tt.input)
		if err != nil {
			t.Fatalf("outputPathForInput(%q): %v", tt.input, err)
		}
		if output != tt.expected {
			t.Errorf("outputPathForInput(%q) = %q, want %q", tt.input, output, tt.expected)
		}
	}
}

func TestIsUnderConverted(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{path: "/Conversions/Converted", expected: true},
		{path: "/Conversions/Converted/MyShow/S01/ep01.mp4", expected: true},
		{path: "/Conversions/MyShow/S01/ep01.mkv", expected: false},
		{path: "/Conversions/movie.mkv", expected: false},
	}

	for _, tt := range tests {
		if got := isUnderConverted(tt.path); got != tt.expected {
			t.Errorf("isUnderConverted(%q) = %v, want %v", tt.path, got, tt.expected)
		}
	}
}

func TestOutputPathForInputOutsideRoot(t *testing.T) {
	_, err := outputPathForInput(filepath.Join(string(filepath.Separator), "tmp", "movie.mkv"))
	if err == nil {
		t.Fatal("expected error for path outside conversions root")
	}
}
