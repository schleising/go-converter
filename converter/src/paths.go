package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const (
	conversionsRoot = "/Conversions"
	convertedRoot   = "/Conversions/Converted"
)

func isSupportedVideo(path string) bool {
	return slices.Contains(supportedExtensions, filepath.Ext(path))
}

func outputPathForInput(inputFile string) (string, error) {
	rel, err := filepath.Rel(conversionsRoot, inputFile)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("input file %q is outside conversions root", inputFile)
	}

	outputRel := strings.TrimSuffix(rel, filepath.Ext(rel)) + ".mp4"
	return filepath.Join(convertedRoot, outputRel), nil
}

func isUnderConverted(path string) bool {
	rel, err := filepath.Rel(convertedRoot, path)
	if err != nil {
		return false
	}

	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func discoverVideoFiles() ([]string, error) {
	var files []string

	err := filepath.WalkDir(conversionsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if path == convertedRoot || isUnderConverted(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isSupportedVideo(path) {
			return nil
		}

		outputFile, err := outputPathForInput(path)
		if err != nil {
			return err
		}

		if _, err := os.Stat(outputFile); err == nil {
			return nil
		}

		files = append(files, path)
		return nil
	})

	return files, err
}
