package bpool

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferPool(t *testing.T) {
	allocSize := 64 * 1024

	pool := New(WithAllocSize(allocSize), WithMaxSize(2*allocSize))
	assert.Equal(t, 0, pool.getLen())

	buf1 := pool.Get(allocSize)
	assert.Equal(t, allocSize, len(buf1))
	assert.Equal(t, allocSize, cap(buf1))
	assert.Equal(t, 0, pool.getLen())

	buf2 := pool.Get(allocSize - 1)
	assert.Equal(t, allocSize, len(buf2))
	assert.Equal(t, allocSize, cap(buf2))
	assert.Equal(t, 0, pool.getLen())

	buf3 := pool.Get(allocSize + 1)
	assert.Equal(t, allocSize+1, len(buf3))
	assert.Equal(t, allocSize+1, cap(buf3))
	assert.Equal(t, 0, pool.getLen())

	buf4 := pool.Get(allocSize)
	assert.Equal(t, allocSize, len(buf4))
	assert.Equal(t, allocSize, cap(buf4))
	assert.Equal(t, 0, pool.getLen())

	pool.Put(buf1)
	assert.Equal(t, 1, pool.getLen())
	assert.Equal(t, allocSize, pool.getSize())

	pool.Put(buf2)
	assert.Equal(t, 2, pool.getLen())
	assert.Equal(t, 2*allocSize, pool.getSize())

	pool.Put(buf3)
	assert.Equal(t, 2, pool.getLen())
	assert.Equal(t, 2*allocSize, pool.getSize())

	pool.Put(buf4)
	assert.Equal(t, 2, pool.getLen())
	assert.Equal(t, 2*allocSize, pool.getSize())

	buf5 := pool.Get(allocSize)
	assert.True(t, testSliceSameAddress(buf2, buf5))

	buf6 := pool.Get(allocSize)
	assert.True(t, testSliceSameAddress(buf1, buf6))

}

func testSliceSameAddress(s1 []byte, s2 []byte) bool {
	return reflect.ValueOf(s1).Pointer() == reflect.ValueOf(s2).Pointer()
}
