package broadcast

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/Joseph-Maulin/go-yt-inference-engine/pkg/services/stream"
)

const (
	startingBroadcastPort   = 9000
	maximumBroadcastStreams = 10
)

type Broadcast struct {
	YoutubeURL string
	Port       int
	cancel     context.CancelFunc
	ctx        context.Context
}

type BroadcastService struct {
	StreamService *stream.StreamService

	broadcasts map[string]*Broadcast
	mu         sync.RWMutex
}

// Lock the BroadcastService mutex before calling this function
func (b *BroadcastService) getNextPort() (int, error) {
	if len(b.broadcasts) >= maximumBroadcastStreams+1 {
		return 0, fmt.Errorf("maximum number of broadcasts reached")
	}

	// Find available port in range starting from port
	portFound := false
	var port int
	for port = startingBroadcastPort; port < startingBroadcastPort+maximumBroadcastStreams; port++ {
		conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
		if err != nil {
			continue
		}
		conn.Close()
		portFound = true
		break
	}
	if !portFound {
		return 0, fmt.Errorf("failed to find a free port")
	}
	return port, nil
}

func NewBroadcastService(streamService *stream.StreamService) *BroadcastService {
	return &BroadcastService{
		StreamService: streamService,
		broadcasts:    make(map[string]*Broadcast),
		mu:            sync.RWMutex{},
	}
}

func (b *BroadcastService) broadcastLoop(broadcast *Broadcast) error {
	// Get the M3U8 URL from the YouTube stream
	b.mu.RLock()
	ytStream, exists := b.StreamService.YouTubeStreams[broadcast.YoutubeURL]
	b.mu.RUnlock()

	if !exists {
		return fmt.Errorf("failed to get YouTube stream for %s", broadcast.YoutubeURL)
	}
	m3u8URL := ytStream.M3U8StreamURL

	log.Printf("Starting direct M3U8->UDP stream for %s", broadcast.YoutubeURL)

	// Use ffmpeg to directly transcode M3U8 to UDP (much faster than GoCV pipeline)
	// This bypasses frame processing but ensures smooth playback
	ffmpegCmd := exec.Command("ffmpeg",
		"-i", m3u8URL,
		"-c:v", "copy", // Copy video codec (no re-encoding!)
		"-c:a", "copy", // Copy audio codec if present
		"-f", "mpegts",
		"-flush_packets", "1",
		"-fflags", "nobuffer",
		fmt.Sprintf("udp://127.0.0.1:%d?pkt_size=1316", broadcast.Port),
	)

	// Capture stderr for debugging
	ffmpegCmd.Stderr = os.Stderr

	if err := ffmpegCmd.Start(); err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	defer func() {
		log.Printf("Cleaning up ffmpeg for %s", broadcast.YoutubeURL)
		if ffmpegCmd.Process != nil {
			ffmpegCmd.Process.Kill()
		}
		ffmpegCmd.Wait()
		time.Sleep(100 * time.Millisecond)
		log.Printf("Cleaned up ffmpeg for %s", broadcast.YoutubeURL)
	}()

	// Wait for context cancellation
	<-broadcast.ctx.Done()
	if err := b.StreamService.RemoveYouTubeStream(broadcast.YoutubeURL); err != nil {
		return fmt.Errorf("failed to stop YouTube stream: %w", err)
	}
	log.Printf("Stopped broadcast for %s on port %d", broadcast.YoutubeURL, broadcast.Port)
	return nil
}

func (b *BroadcastService) StartBroadcast(youtubeURL string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, ok := b.broadcasts[youtubeURL]; ok {
		return fmt.Errorf("broadcast already exists: %s", youtubeURL)
	}

	port, err := b.getNextPort()
	if err != nil {
		return fmt.Errorf("start broadcast failed to get next port: %w", err)
	}

	if err := b.StreamService.AddYouTubeStream(youtubeURL); err != nil {
		return fmt.Errorf("start broadcast failed to start YouTube stream: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	broadcast := &Broadcast{
		YoutubeURL: youtubeURL,
		Port:       port,
		cancel:     cancel,
		ctx:        ctx,
	}
	b.broadcasts[youtubeURL] = broadcast

	go b.broadcastLoop(broadcast)
	log.Printf("Broadcast started; youtubeURL=%s, port=%d\n", youtubeURL, port)
	return nil
}

func (b *BroadcastService) StopBroadcast(youtubeURL string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	broadcast, ok := b.broadcasts[youtubeURL]
	if !ok {
		return fmt.Errorf("broadcast not found; youtubeURL=%s", youtubeURL)
	}
	broadcast.cancel()
	delete(b.broadcasts, youtubeURL)
	log.Printf("Broadcast stopped; youtubeURL=%s\n", youtubeURL)
	return nil
}

func (b *BroadcastService) GetBroadcast(youtubeURL string) (*Broadcast, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	broadcast, ok := b.broadcasts[youtubeURL]
	if !ok {
		return nil, fmt.Errorf("broadcast not found; youtubeURL=%s", youtubeURL)
	}
	return broadcast, nil
}
func (b *BroadcastService) StopAllActiveBroadcasts() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	for _, broadcast := range b.broadcasts {
		broadcast.cancel()
	}
	b.broadcasts = make(map[string]*Broadcast)
	return nil
}

func (b *BroadcastService) Close() {
	b.StopAllActiveBroadcasts()
	log.Println("Closed broadcast service")
}
