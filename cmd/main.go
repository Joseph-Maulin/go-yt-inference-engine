package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Joseph-Maulin/go-yt-inference-engine/pkg/services/broadcast"
	"github.com/Joseph-Maulin/go-yt-inference-engine/pkg/services/stream"
)

var (
	testYouTubeStreamURLs = [2]string{
		"https://www.youtube.com/watch?v=0jEa0FOw6vE",
		"https://www.youtube.com/watch?v=XDThHUawq6E",
	}
)

func main() {

	log.Println("Starting go-yt-inference-engine")

	streamService := stream.NewStream()
	broadcastService := broadcast.NewBroadcastService(streamService)
	defer broadcastService.Close()

	for _, url := range testYouTubeStreamURLs {
		log.Printf("Starting broadcast for: %s", url)
		if err := broadcastService.StartBroadcast(url); err != nil {
			log.Fatalf("Failed to start broadcast for URL=%s: %v", url, err)
		}

		// Log which port was assigned
		if broadcast, err := broadcastService.GetBroadcast(url); err == nil {
			log.Printf("Broadcast started on port %d for %s", broadcast.Port, url)
		}
	}

	log.Println("All broadcasts started successfully")
	log.Println("Press Ctrl+C to stop...")

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down go-yt-inference-engine...")

	log.Println("Shutdown complete")
}
