package telemetry

type streamBufferPoolBuffer struct {
	buf []byte
}

type streamBufferPool struct {
	allocSize int
	buffers   []streamBufferPoolBuffer
}

func newStreamBufferPool(allocSize int) *streamBufferPool {
	return &streamBufferPool{
		allocSize: allocSize,
		buffers:   []streamBufferPoolBuffer{},
	}
}

func (p *streamBufferPool) get(requiredSize int) []byte {
	if requiredSize > p.allocSize {
		return make([]byte, requiredSize)
	}
	if len(p.buffers) == 0 {
		return make([]byte, p.allocSize)
	}
	b := p.buffers[len(p.buffers)-1]
	p.buffers[len(p.buffers)-1] = streamBufferPoolBuffer{}
	p.buffers = p.buffers[:len(p.buffers)-1]
	return b.buf
}

func (p *streamBufferPool) put(buf []byte) {
	if len(buf) != p.allocSize {
		return
	}
	p.buffers = append(p.buffers, streamBufferPoolBuffer{buf: buf})
}

func (p *streamBufferPool) clean(maxSize int) {
	currentSize := p.allocSize * len(p.buffers)
	for currentSize > maxSize && len(p.buffers) > 0 {
		p.buffers[len(p.buffers)-1] = streamBufferPoolBuffer{}
		p.buffers = p.buffers[:len(p.buffers)-1]
		currentSize -= p.allocSize
	}
}

func (p *streamBufferPool) len() int {
	return len(p.buffers)
}
