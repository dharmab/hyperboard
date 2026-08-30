package media

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessVideoRejectsInvalidMedia(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not available")
	}

	_, _, err := ProcessVideo([]byte("not a video"))
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidMedia)
}
