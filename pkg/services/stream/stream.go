package stream

import (
	"context"
	"fmt"
	"sync"

	"log"

	yt "github.com/Joseph-Maulin/go-yt-inference-engine/pkg/services/yt"
	"gocv.io/x/gocv"
)

type StreamService struct {
	YouTubeStreams map[string]*yt.YouTubeStream
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
}

func NewStream() *StreamService {
	ctx, cancel := context.WithCancel(context.Background())
	return &StreamService{
		YouTubeStreams: make(map[string]*yt.YouTubeStream),
		mu:             sync.RWMutex{},
		ctx:            ctx,
		cancel:         cancel,
	}
}

func (s *StreamService) StartYouTubeStream(youtubeURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.YouTubeStreams[youtubeURL]; ok {
		log.Printf("YouTube stream already exists: %s", youtubeURL)
		return fmt.Errorf("YouTube stream=%s already exists", youtubeURL)
	}

	ytStream, err := yt.NewYouTubeStream(youtubeURL)
	if err != nil {
		return fmt.Errorf("failed to create YouTube stream=%s: %w", youtubeURL, err)
	}

	if err := ytStream.Start(); err != nil {
		return fmt.Errorf("failed to start YouTube stream=%s: %w", youtubeURL, err)
	}

	s.YouTubeStreams[youtubeURL] = ytStream
	return nil
}

func (s *StreamService) StopYouTubeStream(youtubeURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ytStream, ok := s.YouTubeStreams[youtubeURL]
	if !ok {
		log.Printf("YouTube stream not found: %s", youtubeURL)
		return fmt.Errorf("YouTube stream=%s not found", youtubeURL)
	}

	ytStream.Stop()
	delete(s.YouTubeStreams, youtubeURL)
	return nil
}

func (s *StreamService) GetFrameChannel(youtubeURL string) (<-chan *gocv.Mat, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ytStream, ok := s.YouTubeStreams[youtubeURL]
	if !ok {
		log.Printf("YouTube stream not found: %s", youtubeURL)
		return nil, fmt.Errorf("YouTube stream=%s not found", youtubeURL)
	}

	return ytStream.GetFrameChan(), nil
}

func (s *StreamService) StopAllYouTubeStreams() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ytStream := range s.YouTubeStreams {
		log.Printf("Stopping YouTube stream: %s", ytStream.YouTubeURL)
		ytStream.Stop()
	}

	s.YouTubeStreams = make(map[string]*yt.YouTubeStream)
	log.Println("Stopped all YouTube streams")
}

func (s *StreamService) Close() {
	s.StopAllYouTubeStreams()
	s.cancel()
	log.Println("Closed stream service")
}
