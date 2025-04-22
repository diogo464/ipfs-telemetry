package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	stream := New()

	var err error
	var segments []Segment

	err = stream.Write(make([]byte, 10))
	segments = stream.Segments(0, 1000)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(segments))
	assert.Equal(t, 22, stream.activeBufferSize)

	err = stream.Write(make([]byte, 100))
	assert.Nil(t, err)
	assert.Equal(t, 0, stream.activeBufferSegStart)
	assert.Equal(t, 134, stream.activeBufferSize)

	stream.addSegment()
	segments = stream.Segments(0, 1000)
	assert.Equal(t, 1, len(segments))
	assert.Equal(t, 134, stream.activeBufferSegStart)
	assert.Equal(t, 134, stream.activeBufferSize)
	assert.Equal(t, 1, len(segments))
	assert.Equal(t, 134, len(segments[0].Data))
	assert.Equal(t, 0, segments[0].SeqN)

	err = stream.Write(make([]byte, 100))
	assert.Nil(t, err)
	stream.addSegment()
	segments = stream.Segments(0, 1000)
	assert.Equal(t, 2, len(segments))
	assert.Equal(t, 134, len(segments[0].Data))
	assert.Equal(t, 0, segments[0].SeqN)
	assert.Equal(t, 112, len(segments[1].Data))
	assert.Equal(t, 1, segments[1].SeqN)
}
