package yt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testYouTubeURL = "https://www.youtube.com/watch?v=XDThHUawq6E"
)

func TestGetM3U8StreamURL(t *testing.T) {
	ytStream, err := NewYouTubeStream(testYouTubeURL)

	assert.NoError(t, err)
	assert.NotNil(t, ytStream)
	assert.NotEmpty(t, ytStream.M3U8StreamURL)
	assert.True(t, strings.HasPrefix(ytStream.M3U8StreamURL, "http"))
}
