package telemetry

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStreamBufferPool(t *testing.T) {
	allocSize := 64 * 1024

	pool := newStreamBufferPool(allocSize)
	assert.Equal(t, 0, pool.len())

	buf1 := pool.get(allocSize)
	assert.Equal(t, allocSize, len(buf1))
	assert.Equal(t, allocSize, cap(buf1))
	assert.Equal(t, 0, pool.len())

	buf2 := pool.get(allocSize - 1)
	assert.Equal(t, allocSize, len(buf2))
	assert.Equal(t, allocSize, cap(buf2))
	assert.Equal(t, 0, pool.len())

	buf3 := pool.get(allocSize + 1)
	assert.Equal(t, allocSize+1, len(buf3))
	assert.Equal(t, allocSize+1, cap(buf3))
	assert.Equal(t, 0, pool.len())

	pool.put(buf1)
	pool.put(buf2)
	pool.put(buf3)
	assert.Equal(t, 2, pool.len())

	buf4 := pool.get(allocSize)
	buf5 := pool.get(allocSize)
	assert.True(t, testSliceSameAddress(buf2, buf4))
	assert.True(t, testSliceSameAddress(buf1, buf5))
	assert.False(t, testSliceSameAddress(buf2, buf5))
	assert.False(t, testSliceSameAddress(buf1, buf4))
	assert.Equal(t, 0, pool.len())

	pool.put(buf1)
	pool.put(buf2)
	assert.Equal(t, 2, pool.len())

	pool.clean(allocSize + 1)
	assert.Equal(t, 1, pool.len())
	pool.clean(allocSize)
	assert.Equal(t, 1, pool.len())
	pool.clean(allocSize - 1)
	assert.Equal(t, 0, pool.len())
}

func TestStreamBufferPoolPutGet(t *testing.T) {
	allocSize := 1024
	pool := newStreamBufferPool(allocSize)

	b1 := pool.get(allocSize)
	b2 := pool.get(allocSize)
	b3 := pool.get(allocSize)
	b4 := pool.get(allocSize)

	pool.put(b1)
	pool.put(b2)
	pool.put(b3)
	pool.put(b4)
}

func testSliceSameAddress(s1 []byte, s2 []byte) bool {
	return reflect.ValueOf(s1).Pointer() == reflect.ValueOf(s2).Pointer()
}
