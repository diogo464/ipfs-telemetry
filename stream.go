package telemetry

import (
	"bytes"
	"encoding/binary"
	"io"
	"sync"
	"time"

	"github.com/diogo464/telemetry/rle"
	"github.com/diogo464/telemetry/vecdeque"
)

type StreamDecoder[T any] func([]byte) (T, error)

type streamDebug struct {
	usedSize  uint32
	totalSize uint32
}

type streamSegmentEntry struct {
	seqN        int
	createdTime time.Time

	data       []byte // slice of the buffer that this segment uses
	buffer     []byte // backing buffer of this segment
	bufferFree bool   // should the buffer be freed when the segment gets cleaned
}

type StreamSegment struct {
	SeqN int
	Data []byte
}

type StreamMessage[T any] struct {
	Timestamp time.Time
	Value     T
}

// Sequence of RLE messages with a lifetime
type Stream struct {
	mu   sync.Mutex
	opts *streamOptions

	segments              *vecdeque.VecDeque[streamSegmentEntry]
	segmentsTotalUsedSize int
	segmentNextSeqN       int
	segmentLastAddTime    time.Time

	activeBuffer         []byte
	activeBufferSize     int
	activeBufferSegStart int

	bufferPool *streamBufferPool
}

func NewStream(o ...StreamOption) *Stream {
	opts := streamDefault()
	streamApply(opts, o...)

	bufferPool := newStreamBufferPool(opts.defaultBufferSize)
	return &Stream{
		opts: opts,

		segments:              vecdeque.New[streamSegmentEntry](),
		segmentsTotalUsedSize: 0,
		segmentNextSeqN:       0,
		segmentLastAddTime:    time.Now(),

		activeBuffer:         bufferPool.get(opts.defaultBufferSize),
		activeBufferSize:     0,
		activeBufferSegStart: 0,

		bufferPool: bufferPool,
	}
}

func (s *Stream) Write(data []byte) error {
	return s.AllocAndWrite(len(data), func(buf []byte) error {
		copy(buf, data)
		return nil
	})
}

func (s *Stream) WriteWithTimestamp(timestamp uint64, data []byte) error {
	return s.AllocAndWriteWithTimestamp(len(data), timestamp, func(buf []byte) error {
		copy(buf, data)
		return nil
	})
}

func (s *Stream) AllocAndWrite(size int, write func([]byte) error) error {
	return s.AllocAndWriteWithTimestamp(size, TimestampNow(), write)
}

func (s *Stream) AllocAndWriteWithTimestamp(size int, timestamp uint64, write func([]byte) error) error {
	const HEADER_SIZE = 12 // 4 len + 8 timestamp

	s.mu.Lock()
	defer s.mu.Unlock()

	requiredSize := size + HEADER_SIZE
	availableSize := len(s.activeBuffer) - s.activeBufferSize
	requiresNewBuffer := requiredSize > availableSize

	if requiredSize > s.opts.maxWriteSize {
		return io.ErrShortWrite
	}

	if s.activeBufferSize > s.activeBufferSegStart && (time.Since(s.segmentLastAddTime) > s.opts.activeBufferLifetime || requiresNewBuffer) {
		s.cleanUpSegments()
		s.addSegment()
	}

	if requiresNewBuffer {
		if !s.segments.IsEmpty() {
			s.segments.BackRef().bufferFree = true
		} else {
			s.bufferPool.put(s.activeBuffer)
		}

		s.activeBuffer = s.bufferPool.get(requiredSize)
		s.activeBufferSegStart = 0
		s.activeBufferSize = 0
	}

	// write message size
	binary.BigEndian.PutUint32(s.activeBuffer[s.activeBufferSize:s.activeBufferSize+4], uint32(size+8))
	binary.BigEndian.PutUint64(s.activeBuffer[s.activeBufferSize+4:s.activeBufferSize+HEADER_SIZE], timestamp)

	start := s.activeBufferSize + HEADER_SIZE
	end := start + size
	buf := s.activeBuffer[start:end]
	err := write(buf)
	if err == nil {
		s.activeBufferSize += requiredSize
	}

	return err
}

func (s *Stream) Segments(since int, n int) []StreamSegment {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanUpSegments()

	remain := n
	index := 0
	segments := make([]StreamSegment, 0)
	for remain > 0 && index < s.segments.Len() {
		entry := s.segments.Get(index)
		if entry.seqN >= since {
			segments = append(segments, StreamSegment{
				SeqN: entry.seqN,
				Data: entry.data,
			})
			remain -= 1
		}
		index += 1
	}
	return segments
}

func (s *Stream) LatestSeqN() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.segmentNextSeqN - 1
}

func (s *Stream) addSegment() {
	now := time.Now()
	segmentData := s.activeBuffer[s.activeBufferSegStart:s.activeBufferSize]
	s.segments.PushBack(streamSegmentEntry{
		seqN:        s.segmentNextSeqN,
		createdTime: now,
		data:        segmentData,
		buffer:      s.activeBuffer,
		bufferFree:  false,
	})
	s.segmentNextSeqN += 1
	s.segmentsTotalUsedSize += len(segmentData)
	s.segmentLastAddTime = now
	s.activeBufferSegStart = s.activeBufferSize
}

func (s *Stream) cleanUpSegments() {
	for !s.segments.IsEmpty() && (s.segmentsTotalUsedSize > s.opts.maxSize || time.Since(s.segments.Front().createdTime) > s.opts.segmentLifetime) {
		entry := s.segments.PopFront()
		s.segmentsTotalUsedSize -= len(entry.data)
		if entry.bufferFree {
			s.bufferPool.put(entry.buffer)
		}
	}
	s.bufferPool.clean(s.opts.maxSize - s.segmentsTotalUsedSize)
}

func (s *Stream) debug() streamDebug {
	s.mu.Lock()
	defer s.mu.Unlock()

	var usedSize uint32 = 0
	var totalSize uint32 = 0

	for i := 0; i < s.segments.Len(); i++ {
		segment := s.segments.Get(i)
		usedSize += uint32(len(segment.data))
		if segment.bufferFree {
			totalSize += uint32(len(segment.buffer))
		}
	}
	usedSize += uint32(s.activeBufferSize - s.activeBufferSegStart)
	totalSize += uint32(len(s.activeBuffer))

	return streamDebug{
		usedSize:  usedSize,
		totalSize: totalSize,
	}
}

func StreamSegmentDecode[T any](decoder StreamDecoder[T], segment StreamSegment) ([]StreamMessage[T], error) {
	items := make([]StreamMessage[T], 0)
	reader := bytes.NewReader(segment.Data)
	for {
		buf, err := rle.Read(reader)
		if err == io.EOF {
			break
		}
		if len(buf) < 8 {
			return nil, io.ErrUnexpectedEOF
		}
		timestamp := binary.BigEndian.Uint64(buf[:8])
		item, err := decoder(buf[8:])
		if err != nil {
			return nil, err
		}
		items = append(items, StreamMessage[T]{
			Timestamp: time.Unix(int64(timestamp/1_000_000_000), int64(timestamp%1_000_000_000)),
			Value:     item,
		})
	}
	return items, nil
}
