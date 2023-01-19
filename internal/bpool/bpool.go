package bpool

import "sync"

type buffer []byte

type Pool struct {
	mu      sync.Mutex
	opts    *options
	size    int
	buffers []buffer
}

func New(opts ...Option) *Pool {
	options := defaults()
	for _, opt := range opts {
		opt(options)
	}

	return &Pool{
		opts:    options,
		size:    0,
		buffers: make([]buffer, 0),
	}
}

func (p *Pool) Get(size int) buffer {
	p.mu.Lock()
	defer p.mu.Unlock()

	if size > p.opts.allocSize {
		return make(buffer, size)
	}

	if len(p.buffers) == 0 {
		return make(buffer, p.opts.allocSize)
	}

	b := p.buffers[len(p.buffers)-1]
	p.buffers[len(p.buffers)-1] = nil
	p.buffers = p.buffers[:len(p.buffers)-1]
	p.size -= cap(b)

	return b
}

func (p *Pool) Put(b buffer) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(b) != p.opts.allocSize {
		return
	}

	if p.size+len(b) > p.opts.maxSize {
		return
	}

	p.buffers = append(p.buffers, b)
	p.size += cap(b)
}

func (p *Pool) AllocSize() int {
	return p.opts.allocSize
}

func (p *Pool) getSize() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.size
}

func (p *Pool) getLen() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.buffers)
}
