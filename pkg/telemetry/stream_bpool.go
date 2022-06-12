package telemetry

type streamBufferPoolBuffer struct {
	buf []byte
}

type streamBufferPool struct {
	allocSize int

	buffers     []streamBufferPoolBuffer
	buffersSize int
}

func newStreamBufferPool(allocSize int) *streamBufferPool {
	return &streamBufferPool{
		allocSize: allocSize,

		buffers:     []streamBufferPoolBuffer{},
		buffersSize: 0,
	}
}

func (p *streamBufferPool) get(requiredSize int) []byte {
	if requiredSize > p.allocSize {
		return make([]byte, requiredSize)
	}
	if p.buffersSize > 0 {
		b := p.buffers[p.buffersSize-1]
		p.buffers[p.buffersSize-1] = streamBufferPoolBuffer{}
		p.buffersSize -= 1
		return b.buf
	}
	return make([]byte, p.allocSize)
}

func (p *streamBufferPool) put(buf []byte) {
	if len(buf) != p.allocSize {
		return
	}

	if p.buffersSize == cap(p.buffers) {
		p.buffers = append(p.buffers, streamBufferPoolBuffer{buf: buf})
	} else {
		p.buffers[p.buffersSize] = streamBufferPoolBuffer{buf: buf}
	}
	p.buffersSize += 1
}

func (p *streamBufferPool) clean(maxSize int) {
	currentSize := p.allocSize * p.buffersSize
	for currentSize > maxSize && p.buffersSize > 0 {
		p.buffers[p.buffersSize-1] = streamBufferPoolBuffer{}
		p.buffersSize -= 1
		currentSize -= p.allocSize
	}
}

func (p *streamBufferPool) len() int {
	return p.buffersSize
}
