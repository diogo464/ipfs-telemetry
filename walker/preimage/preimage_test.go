package preimage

import (
	"strconv"
	"testing"

	kbucket "github.com/libp2p/go-libp2p-kbucket"
)

func TestPrefixFromHash(t *testing.T) {
	hash := [32]byte{0b1010_1100, 0b0000_1111, 0b0000_1001}

	prefix24 := prefixFromHash(24, hash)
	expected24 := 0b1010_1100_0000_1111_0000_1001
	assertPrefix(t, 24, prefix24, expected24)

	prefix23 := prefixFromHash(23, hash)
	expected23 := 0b1010_1100_0000_1111_0000_100
	assertPrefix(t, 23, prefix23, expected23)

	prefix16 := prefixFromHash(16, hash)
	expected16 := 0b1010_1100_0000_1111
	assertPrefix(t, 16, prefix16, expected16)

	prefix15 := prefixFromHash(15, hash)
	expected15 := 0b1010_1100_0000_111
	assertPrefix(t, 15, prefix15, expected15)

	prefix8 := prefixFromHash(8, hash)
	expected8 := 0b1010_1100
	assertPrefix(t, 8, prefix8, expected8)

	prefix1 := prefixFromHash(1, hash)
	expected1 := 0b1
	assertPrefix(t, 1, prefix1, expected1)

	prefix0 := prefixFromHash(0, hash)
	expected0 := 0
	assertPrefix(t, 0, prefix0, expected0)
}

func TestMaskedPrefixFromHash(t *testing.T) {
	hash := [32]byte{0b1010_1100, 0b0000_1111, 0b0000_1001}

	prefix24 := maskedPrefixFromHash(24, hash, 0)
	expected24 := 0b0010_1100_0000_1111_0000_1001
	assertPrefix(t, 24, prefix24, expected24)

	prefix23 := maskedPrefixFromHash(24, hash, 23)
	expected23 := 0b1010_1100_0000_1111_0000_1000
	assertPrefix(t, 1, prefix23, expected23)

	prefix1 := maskedPrefixFromHash(24, hash, 1)
	expected1 := 0b1110_1100_0000_1111_0000_1001
	assertPrefix(t, 1, prefix1, expected1)
}

func assertPrefix(t *testing.T, nbits int, prefix int, expected int) {
	if prefix != expected {
		t.Fatalf(
			"Invalid %v bit prefix\nComputed: %v\nExpected: %v\n",
			nbits,
			strconv.FormatInt(int64(prefix), 2),
			strconv.FormatInt(int64(expected), 2),
		)
	}
}

func TestBucketPlacement(t *testing.T) {
	pid := generateRandomPeerID()
	tmpRTID := kbucket.ConvertPeerID(pid)
	table := Generate()
	ids := table.GetIDsForPeer(pid)
	for i, id := range ids {
		cpl := kbucket.CommonPrefixLen(tmpRTID, kbucket.ConvertPeerID(id))
		if cpl != i {
			t.Error("Invalid CPL, got ", cpl, " expected ", i)
		}
	}
}
