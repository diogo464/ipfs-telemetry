package snapshot

import (
	"testing"
	"time"
)

func TestWindow(t *testing.T) {
	w := newWindowImpl(time.Second)
	w.Push(NewSnapshot("tag0", nil)) // seqn 1
	w.Push(NewSnapshot("tag1", nil)) // seqn 2
	w.Push(NewSnapshot("tag2", nil)) // seqn 3
	w.Push(NewSnapshot("tag2", nil)) // seqn 4
	w.Push(NewSnapshot("tag3", nil)) // seqn 5

	{
		s := w.Since(1)
		if len(s) != 5 {
			t.Fatalf("invalid length, got %v expected 5", len(s))
		}
	}

	{
		s := w.Since(2)
		if len(s) != 4 {
			t.Fatalf("invalid length, got %v expected 4", len(s))
		}
	}

	{
		s := w.Since(4)
		if len(s) != 2 {
			t.Fatalf("invalid length, got %v expected 2", len(s))
		}
	}

	{
		s := w.Since(6)
		if len(s) != 0 {
			t.Fatalf("invalid length, got %v expected 0", len(s))
		}
	}
}
