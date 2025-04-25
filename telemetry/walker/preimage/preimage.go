package preimage

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"
	"runtime"

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multihash"
)

const (
	MaxBits       = 24
	DefaultBits   = 16
	invalidPeerID = peer.ID("")
)

type prefixedPeerID struct {
	ID     peer.ID
	Prefix int
}

type tableJson struct {
	Bits int       `json:"bits"`
	Pids []peer.ID `json:"pids"`
}

type Table struct {
	// number of bits
	bits int
	pids []peer.ID
}

func prefixFromHash(bits int, hash [32]byte) int {
	prefix := int(hash[0])<<16 | int(hash[1])<<8 | int(hash[2])
	prefix = prefix >> (MaxBits - bits)
	return prefix
}

func maskedPrefixFromHash(bits int, hash [32]byte, offset int) int {
	mask := 1 << (bits - offset - 1)
	return prefixFromHash(bits, hash) ^ mask
}

func (t *Table) UnmarshalJSON(data []byte) error {
	tablej := tableJson{}
	if err := json.Unmarshal(data, &tablej); err != nil {
		return err
	}
	t.bits = tablej.Bits
	t.pids = tablej.Pids
	return nil
}

func (t *Table) MarshalJSON() ([]byte, error) {
	tablej := tableJson{Bits: t.bits, Pids: t.pids}
	return json.Marshal(&tablej)
}

func (t *Table) UnmarshalBinary(compressed []byte) error {
	freader := flate.NewReader(bytes.NewReader(compressed))
	marshaled, err := io.ReadAll(freader)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshaled, t)
}

func (t *Table) MarshalBinary() ([]byte, error) {
	marshaled, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	writer := bytes.Buffer{}
	fwriter, err := flate.NewWriter(&writer, flate.DefaultCompression)
	if err != nil {
		return nil, err
	}
	_, err = fwriter.Write(marshaled)
	if err != nil {
		return nil, err
	}
	err = fwriter.Flush()
	if err != nil {
		return nil, err
	}
	return writer.Bytes(), nil
}

func (t *Table) GetIDsForPeer(p peer.ID) []peer.ID {
	ids := make([]peer.ID, 0, t.bits)
	hash := sha256.Sum256([]byte(p))

	for bit := 0; bit < t.bits; bit++ {
		prefix := maskedPrefixFromHash(t.bits, hash, bit)
		ids = append(ids, t.pids[prefix])
	}

	return ids
}

func (t *Table) SerializeTable() ([]byte, error) {
	return json.Marshal(t.pids)
}

func Generate() *Table {
	return GenerateWithBits(DefaultBits)
}

func GenerateWithBits(bits int) *Table {
	t := make([]peer.ID, 1<<bits)
	for i := range t {
		t[i] = invalidPeerID
	}

	work := make(chan prefixedPeerID)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// worker goroutines
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go func(ctx context.Context) {
		LOOP:
			for {
				p := generateRandomPeerID()
				hash := sha256.Sum256([]byte(p))
				prefix := int(hash[0])<<8 | int(hash[1])
				work <- prefixedPeerID{ID: p, Prefix: prefix}
				select {
				case <-ctx.Done():
					break LOOP
				default:
				}
			}

		}(ctx)
	}

	total := 1 << bits
	missing := total
	for missing > 0 {
		ppid := <-work
		if t[ppid.Prefix] == invalidPeerID {
			t[ppid.Prefix] = ppid.ID
			missing -= 1
		}
	}

	return &Table{pids: t, bits: bits}
}

func generateRandomPeerID() peer.ID {
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	var alg uint64 = multihash.SHA2_256
	hash, _ := multihash.Sum(b, alg, -1)
	return peer.ID(hash)
}
