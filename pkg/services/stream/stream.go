package stream

import (
	"fmt"
	"sync"

	"log"

	yt "github.com/Joseph-Maulin/go-yt-inference-engine/pkg/services/yt"
)

type StreamService struct {
	YouTubeStreams map[string]*yt.YouTubeStream
	mu             sync.RWMutex
}

func NewStream() *StreamService {
	return &StreamService{
		YouTubeStreams: make(map[string]*yt.YouTubeStream),
		mu:             sync.RWMutex{},
	}
}

func (s *StreamService) AddYouTubeStream(youtubeURL string) error {
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

	s.YouTubeStreams[youtubeURL] = ytStream
	return nil
}

func (s *StreamService) RemoveYouTubeStream(youtubeURL string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.YouTubeStreams[youtubeURL]
	if !ok {
		log.Printf("YouTube stream not found: %s", youtubeURL)
		return fmt.Errorf("YouTube stream=%s not found", youtubeURL)
	}

	delete(s.YouTubeStreams, youtubeURL)
	return nil
}
