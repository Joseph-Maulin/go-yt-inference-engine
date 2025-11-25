package yt

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"gocv.io/x/gocv"
)

//go:embed yt_dlp.py
var pythonScript string

type YouTubeStream struct {
	YouTubeURL    string
	M3U8StreamURL string
	FrameChan     chan *gocv.Mat
	stopChan      chan struct{}
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
		FrameChan:     make(chan *gocv.Mat, 100),
		stopChan:      make(chan struct{}),
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

func (y *YouTubeStream) Start() error {
	log.Printf("Starting YouTube stream for URL=%s", y.YouTubeURL)
	capture, err := gocv.OpenVideoCapture(y.M3U8StreamURL)
	if err != nil {
		log.Printf("Failed to open video capture: %v", err)
		return err
	}

	go func() {
		log.Printf("Started YouTube stream for URL=%s", y.YouTubeURL)
		defer capture.Close()
		defer close(y.FrameChan)

		frame := gocv.NewMat()
		defer frame.Close()

		for {
			select {
			case <-y.ctx.Done():
				log.Println("Stopping YouTube stream")
				return
			default:
				if ok := capture.Read(&frame); !ok {
					log.Println("Error reading frame from YouTube stream")
					capture.Close()
					capture, err = gocv.OpenVideoCapture(y.M3U8StreamURL)
					if err != nil {
						log.Printf("Failed to reconnect: %v", err)
						return
					}
					continue
				}

				if frame.Empty() {
					continue
				}

				frameCopy := frame.Clone()

				select {
				case y.FrameChan <- &frameCopy:
					continue
				case <-y.ctx.Done():
					log.Println("Stopping YouTube stream")
					return
				}

			}
		}
	}()

	return nil
}

func (y *YouTubeStream) Stop() {
	log.Printf("Stopping YouTube stream: %s", y.YouTubeURL)
	y.cancel()
	close(y.stopChan)
	log.Printf("YouTube stream stopped successfully: %s", y.YouTubeURL)
}

func (y *YouTubeStream) GetFrameChan() <-chan *gocv.Mat {
	return y.FrameChan
}
