package rle

import (
	"bytes"
	"testing"
)

func TestRead_Empty(t *testing.T) {
	testRead(t,
		[]byte{0, 0, 0, 5, 255, 55, 22, 11, 44},
		[]byte{},
	)
}

func TestRead_NotEmpty(t *testing.T) {
	testRead(t,
		[]byte{0, 0, 0, 5, 255, 55, 22, 11, 44},
		[]byte{255, 55, 22, 11, 44},
	)
}

func TestWrite_Empty(t *testing.T) {
	testWrite(t,
		[]byte{},
		[]byte{0, 0, 0, 0},
	)
}

func TestWrite_NotEmpty(t *testing.T) {
	testWrite(t,
		[]byte{255, 22, 10},
		[]byte{0, 0, 0, 3, 255, 22, 10},
	)
}

func testRead(t *testing.T, buf []byte, expected []byte) {
	reader := bytes.NewReader(buf)
	msg, err := Read(reader)

	if err != nil {
		t.Fatal(err)
	}

	if len(msg) != 5 {
		t.Fatal("message length should be 5")
	}

	for i, v := range expected {
		if msg[i] != v {
			t.Fatalf("message is different at %v, %v != %v", i, msg[i], expected[i])
		}
	}
}

func testWrite(t *testing.T, msg []byte, expected []byte) {
	writer := new(bytes.Buffer)
	if err := Write(writer, msg); err != nil {
		t.Fatal(err)
	}

	written := writer.Bytes()

	if len(written) != len(expected) {
		t.Fatalf("length differs")
	}

	for i, v := range expected {
		if written[i] != v {
			t.Fatalf("message is different at %v, %v != %v", i, written[i], expected[i])
		}
	}
}
