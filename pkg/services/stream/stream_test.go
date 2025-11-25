package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testYouTubeURL = "https://www.youtube.com/watch?v=XDThHUawq6E"
)

func TestStreamService(t *testing.T) {
	streamService := NewStream()
	defer streamService.Close()

	assert.NotNil(t, streamService)

	err := streamService.StartYouTubeStream(testYouTubeURL)
	assert.NoError(t, err)
	defer streamService.StopYouTubeStream(testYouTubeURL)

	frameChan, err := streamService.GetFrameChannel(testYouTubeURL)
	assert.NoError(t, err)
	assert.NotNil(t, frameChan)

	frame := <-frameChan
	assert.NotNil(t, frame)

	assert.False(t, frame.Empty())
	t.Logf("Received frame from YouTube stream; shape=%v, type=%v", frame.Size(), frame.Type())
	frame.Close()
}
