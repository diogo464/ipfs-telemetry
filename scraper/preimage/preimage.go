package preimage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"runtime"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multihash"
)

const PREIMAGETABLE_BITS = 16
const INVALID_PEER_ID = peer.ID("")

type prefixedPeerID struct {
	ID     peer.ID
	Prefix int
}

type Table struct {
	t []peer.ID
}

func (t *Table) GetIDsForPeer(p peer.ID) []peer.ID {
	ids := make([]peer.ID, 0, PREIMAGETABLE_BITS)
	ph := sha256.Sum256([]byte(p))

	for bit := 0; bit < PREIMAGETABLE_BITS; bit++ {
		mask := 1 << (PREIMAGETABLE_BITS - bit - 1)
		prefix := int(ph[0])<<8 | int(ph[1])
		prefix = prefix ^ mask
		ids = append(ids, t.t[prefix])
	}

	return ids
}

func (t *Table) SerializeTable() ([]byte, error) {
	return json.Marshal(t.t)
}

func DeserializeTable(data []byte) (*Table, error) {
	t := []peer.ID{}
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &Table{t: t}, nil
}

func DeserializeTableFromFile(path string) (*Table, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return DeserializeTable(data)
}

func GenerateTable() *Table {
	t := make([]peer.ID, 1<<PREIMAGETABLE_BITS)
	for i := range t {
		t[i] = INVALID_PEER_ID
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

	total := 1 << PREIMAGETABLE_BITS
	missing := total
	for missing > 0 {
		ppid := <-work
		if t[ppid.Prefix] == INVALID_PEER_ID {
			t[ppid.Prefix] = ppid.ID
			missing -= 1
			if missing%10000 == 0 {
				fmt.Printf("%v %%\n", (float64(total-missing)/float64(total))*100)
			}
		}
	}

	return &Table{t: t}
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
