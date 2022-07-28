package telemetry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	stream := NewStream()

	var err error
	var segments []StreamSegment

	err = stream.Write(make([]byte, 10))
	segments = stream.Segments(0, 1000)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(segments))
	assert.Equal(t, 14, stream.activeBufferSize)

	err = stream.Write(make([]byte, 100))
	assert.Nil(t, err)
	assert.Equal(t, 0, stream.activeBufferSegStart)
	assert.Equal(t, 118, stream.activeBufferSize)

	stream.addSegment()
	segments = stream.Segments(0, 1000)
	assert.Equal(t, 1, len(segments))
	assert.Equal(t, 118, stream.activeBufferSegStart)
	assert.Equal(t, 118, stream.activeBufferSize)
	assert.Equal(t, 1, len(segments))
	assert.Equal(t, 118, len(segments[0].Data))
	assert.Equal(t, 0, segments[0].SeqN)

	err = stream.Write(make([]byte, 100))
	assert.Nil(t, err)
	stream.addSegment()
	segments = stream.Segments(0, 1000)
	assert.Equal(t, 2, len(segments))
	assert.Equal(t, 118, len(segments[0].Data))
	assert.Equal(t, 0, segments[0].SeqN)
	assert.Equal(t, 104, len(segments[1].Data))
	assert.Equal(t, 1, segments[1].SeqN)
}
