package broadcast

import (
	"testing"
	"time"

	"github.com/Joseph-Maulin/go-yt-inference-engine/pkg/services/stream"
	"github.com/stretchr/testify/assert"
)

const (
	testYouTubeURL = "https://www.youtube.com/watch?v=0jEa0FOw6vE" // Western wall of Jerusalem
)

func TestBroadcastService(t *testing.T) {

	streamService := stream.NewStream()
	defer streamService.Close()

	broadcastService := NewBroadcastService(streamService)
	defer broadcastService.Close()

	assert.NotNil(t, broadcastService)

	// Start YouTube stream FIRST
	err := streamService.StartYouTubeStream(testYouTubeURL)
	assert.NoError(t, err)

	// Now start the broadcast
	err = broadcastService.StartBroadcast(testYouTubeURL)
	assert.NoError(t, err)

	// Verify broadcast was created
	broadcast, err := broadcastService.GetBroadcast(testYouTubeURL)
	assert.NoError(t, err)
	assert.NotNil(t, broadcast)
	t.Logf("Broadcast started on port %d", broadcast.Port)

	// Let it run for a bit
	t.Log("Broadcast running, waiting 30 seconds...")
	time.Sleep(30 * time.Second)

	// Stop the broadcast
	err = broadcastService.StopBroadcast(testYouTubeURL)
	assert.NoError(t, err)
	t.Log("Broadcast stopped successfully")
}
