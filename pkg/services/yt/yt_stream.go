package yt

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

//go:embed yt_dlp.py
var pythonScript string

type YouTubeStream struct {
	YouTubeURL    string
	M3U8StreamURL string
	ctx           context.Context
	cancel        context.CancelFunc
}

func GetM3U8StreamURLFromYouTubeURL(youTubeURL string) (string, error) {
	// Create a temporary file to write the embedded Python script
	tmpFile, err := os.CreateTemp("", "yt_dlp_*.py")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name()) // Clean up
	defer tmpFile.Close()

	// Write the embedded script to the temp file
	if _, err := tmpFile.WriteString(pythonScript); err != nil {
		return "", err
	}
	tmpFile.Close()

	// Execute the Python script
	cmd := exec.Command("uv", "run", tmpFile.Name(), youTubeURL)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	// Parse http from output
	outputLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(outputLines) == 0 {
		return "", fmt.Errorf("no output from Python script")
	}

	url := strings.TrimSpace(outputLines[len(outputLines)-1])

	if !strings.HasPrefix(url, "http") {
		return "", fmt.Errorf("invalid URL from Python script: %s", url)
	}

	return url, nil
}

func NewYouTubeStream(youTubeURL string) (*YouTubeStream, error) {
	log.Printf("Creating YouTube stream for URL=%s", youTubeURL)

	// Get the m3u8 stream URL
	m3u8StreamURL, err := GetM3U8StreamURLFromYouTubeURL(youTubeURL)
	if err != nil {
		return nil, err
	}

	log.Printf("M3U8 stream URL: %s", m3u8StreamURL)

	ctx, cancel := context.WithCancel(context.Background())

	return &YouTubeStream{
		YouTubeURL:    youTubeURL,
		M3U8StreamURL: m3u8StreamURL,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}
