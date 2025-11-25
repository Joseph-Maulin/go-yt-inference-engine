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

	assert.NotNil(t, streamService)

	err := streamService.AddYouTubeStream(testYouTubeURL)
	assert.NoError(t, err)
	defer streamService.RemoveYouTubeStream(testYouTubeURL)
}
