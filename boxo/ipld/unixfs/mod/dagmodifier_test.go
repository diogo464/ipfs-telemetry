package mod

import (
	"context"
	"fmt"
	"io"
	"testing"

	dag "github.com/ipfs/boxo/ipld/merkledag"
	"github.com/ipfs/boxo/ipld/unixfs"
	trickle "github.com/ipfs/boxo/ipld/unixfs/importer/trickle"
	uio "github.com/ipfs/boxo/ipld/unixfs/io"
	testu "github.com/ipfs/boxo/ipld/unixfs/test"
	"github.com/ipfs/go-test/random"
)

func testModWrite(t *testing.T, beg, size uint64, orig []byte, dm *DagModifier, opts testu.NodeOpts) []byte {
	newdata := make([]byte, size)
	random.NewRand().Read(newdata)

	if size+beg > uint64(len(orig)) {
		orig = append(orig, make([]byte, (size+beg)-uint64(len(orig)))...)
	}
	copy(orig[beg:], newdata)

	nmod, err := dm.WriteAt(newdata, int64(beg))
	if err != nil {
		t.Fatal(err)
	}

	if nmod != int(size) {
		t.Fatalf("Mod length not correct! %d != %d", nmod, size)
	}

	verifyNode(t, orig, dm, opts)

	return orig
}

func verifyNode(t *testing.T, orig []byte, dm *DagModifier, opts testu.NodeOpts) {
	nd, err := dm.GetNode()
	if err != nil {
		t.Fatal(err)
	}

	err = trickle.VerifyTrickleDagStructure(nd, trickle.VerifyParams{
		Getter:      dm.dagserv,
		Direct:      dm.MaxLinks,
		LayerRepeat: 4,
		Prefix:      &opts.Prefix,
		RawLeaves:   opts.RawLeavesUsed,
	})
	if err != nil {
		t.Fatal(err)
	}

	rd, err := uio.NewDagReader(context.Background(), nd, dm.dagserv)
	if err != nil {
		t.Fatal(err)
	}

	after, err := io.ReadAll(rd)
	if err != nil {
		t.Fatal(err)
	}

	err = testu.ArrComp(after, orig)
	if err != nil {
		t.Fatal(err)
	}
}

func runAllSubtests(t *testing.T, tfunc func(*testing.T, testu.NodeOpts)) {
	t.Run("opts=ProtoBufLeaves", func(t *testing.T) { tfunc(t, testu.UseProtoBufLeaves) })
	t.Run("opts=RawLeaves", func(t *testing.T) { tfunc(t, testu.UseRawLeaves) })
	t.Run("opts=CidV1", func(t *testing.T) { tfunc(t, testu.UseCidV1) })
	t.Run("opts=Blake2b256", func(t *testing.T) { tfunc(t, testu.UseBlake2b256) })
}

func TestDagModifierBasic(t *testing.T) {
	runAllSubtests(t, testDagModifierBasic)
}

func testDagModifierBasic(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	b, n := testu.GetRandomNode(t, dserv, 50000, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	// Within zero block
	beg := uint64(15)
	length := uint64(60)

	t.Log("Testing mod within zero block")
	b = testModWrite(t, beg, length, b, dagmod, opts)

	// Within bounds of existing file
	beg = 1000
	length = 4000
	t.Log("Testing mod within bounds of existing multiblock file.")
	b = testModWrite(t, beg, length, b, dagmod, opts)

	// Extend bounds
	beg = 49500
	length = 4000

	t.Log("Testing mod that extends file.")
	b = testModWrite(t, beg, length, b, dagmod, opts)

	// "Append"
	beg = uint64(len(b))
	length = 3000
	t.Log("Testing pure append")
	_ = testModWrite(t, beg, length, b, dagmod, opts)

	// Verify reported length
	node, err := dagmod.GetNode()
	if err != nil {
		t.Fatal(err)
	}

	size, err := fileSize(node)
	if err != nil {
		t.Fatal(err)
	}

	const expected = uint64(50000 + 3500 + 3000)
	if size != expected {
		t.Fatalf("Final reported size is incorrect [%d != %d]", size, expected)
	}
}

func TestMultiWrite(t *testing.T) {
	runAllSubtests(t, testMultiWrite)
}

func testMultiWrite(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	data := make([]byte, 4000)
	random.NewRand().Read(data)

	for i := 0; i < len(data); i++ {
		n, err := dagmod.WriteAt(data[i:i+1], int64(i))
		if err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatal("Somehow wrote the wrong number of bytes! (n != 1)")
		}

		size, err := dagmod.Size()
		if err != nil {
			t.Fatal(err)
		}

		if size != int64(i+1) {
			t.Fatal("Size was reported incorrectly")
		}
	}

	verifyNode(t, data, dagmod, opts)
}

func TestMultiWriteAndFlush(t *testing.T) {
	runAllSubtests(t, testMultiWriteAndFlush)
}

func testMultiWriteAndFlush(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	data := make([]byte, 20)
	random.NewRand().Read(data)

	for i := 0; i < len(data); i++ {
		n, err := dagmod.WriteAt(data[i:i+1], int64(i))
		if err != nil {
			t.Fatal(err)
		}
		if n != 1 {
			t.Fatal("Somehow wrote the wrong number of bytes! (n != 1)")
		}
		err = dagmod.Sync()
		if err != nil {
			t.Fatal(err)
		}
	}

	verifyNode(t, data, dagmod, opts)
}

func TestWriteNewFile(t *testing.T) {
	runAllSubtests(t, testWriteNewFile)
}

func testWriteNewFile(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	towrite := make([]byte, 2000)
	random.NewRand().Read(towrite)

	nw, err := dagmod.Write(towrite)
	if err != nil {
		t.Fatal(err)
	}
	if nw != len(towrite) {
		t.Fatal("Wrote wrong amount")
	}

	verifyNode(t, towrite, dagmod, opts)
}

func TestMultiWriteCoal(t *testing.T) {
	runAllSubtests(t, testMultiWriteCoal)
}

func testMultiWriteCoal(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	data := make([]byte, 1000)
	random.NewRand().Read(data)

	for i := 0; i < len(data); i++ {
		n, err := dagmod.WriteAt(data[:i+1], 0)
		if err != nil {
			fmt.Println("FAIL AT ", i)
			t.Fatal(err)
		}
		if n != i+1 {
			t.Fatal("Somehow wrote the wrong number of bytes! (n != 1)")
		}

	}

	verifyNode(t, data, dagmod, opts)
}

func TestLargeWriteChunks(t *testing.T) {
	runAllSubtests(t, testLargeWriteChunks)
}

func testLargeWriteChunks(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	const wrsize = 1000
	const datasize = 10000000
	data := make([]byte, datasize)

	random.NewRand().Read(data)

	for i := 0; i < datasize/wrsize; i++ {
		n, err := dagmod.WriteAt(data[i*wrsize:(i+1)*wrsize], int64(i*wrsize))
		if err != nil {
			t.Fatal(err)
		}
		if n != wrsize {
			t.Fatal("failed to write buffer")
		}
	}

	_, err = dagmod.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(dagmod)
	if err != nil {
		t.Fatal(err)
	}

	if err = testu.ArrComp(out, data); err != nil {
		t.Fatal(err)
	}
}

func TestDagTruncate(t *testing.T) {
	runAllSubtests(t, testDagTruncate)
}

func testDagTruncate(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	b, n := testu.GetRandomNode(t, dserv, 50000, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	err = dagmod.Truncate(12345)
	if err != nil {
		t.Fatal(err)
	}
	size, err := dagmod.Size()
	if err != nil {
		t.Fatal(err)
	}

	if size != 12345 {
		t.Fatal("size was incorrect!")
	}

	_, err = dagmod.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(dagmod)
	if err != nil {
		t.Fatal(err)
	}

	if err = testu.ArrComp(out, b[:12345]); err != nil {
		t.Fatal(err)
	}

	err = dagmod.Truncate(10)
	if err != nil {
		t.Fatal(err)
	}

	size, err = dagmod.Size()
	if err != nil {
		t.Fatal(err)
	}

	if size != 10 {
		t.Fatal("size was incorrect!")
	}

	err = dagmod.Truncate(0)
	if err != nil {
		t.Fatal(err)
	}

	size, err = dagmod.Size()
	if err != nil {
		t.Fatal(err)
	}

	if size != 0 {
		t.Fatal("size was incorrect!")
	}
}

// TestDagSync tests that a DAG will expand sparse during sync
// if offset > curNode's size.
func TestDagSync(t *testing.T) {
	dserv := testu.GetDAGServ()
	nd := dag.NodeWithData(unixfs.FilePBData(nil, 0))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, nd, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}

	_, err = dagmod.Write([]byte("test1"))
	if err != nil {
		t.Fatal(err)
	}

	err = dagmod.Sync()
	if err != nil {
		t.Fatal(err)
	}

	// Truncate leave the offset at 5 and filesize at 0
	err = dagmod.Truncate(0)
	if err != nil {
		t.Fatal(err)
	}

	_, err = dagmod.Write([]byte("test2"))
	if err != nil {
		t.Fatal(err)
	}

	// When Offset > filesize , Sync will call enpandSparse
	err = dagmod.Sync()
	if err != nil {
		t.Fatal(err)
	}

	_, err = dagmod.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(dagmod)
	if err != nil {
		t.Fatal(err)
	}

	if err = testu.ArrComp(out[5:], []byte("test2")); err != nil {
		t.Fatal(err)
	}
}

// TestDagTruncateSameSize tests that a DAG truncated
// to the same size (i.e., doing nothing) doesn't modify
// the DAG (its hash).
func TestDagTruncateSameSize(t *testing.T) {
	runAllSubtests(t, testDagTruncateSameSize)
}

func testDagTruncateSameSize(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	_, n := testu.GetRandomNode(t, dserv, 50000, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	// Copied from `TestDagTruncate`.

	size, err := dagmod.Size()
	if err != nil {
		t.Fatal(err)
	}

	err = dagmod.Truncate(size)
	if err != nil {
		t.Fatal(err)
	}

	modifiedNode, err := dagmod.GetNode()
	if err != nil {
		t.Fatal(err)
	}

	if modifiedNode.Cid().Equals(n.Cid()) == false {
		t.Fatal("the node has been modified!")
	}
}

func TestSparseWrite(t *testing.T) {
	runAllSubtests(t, testSparseWrite)
}

func testSparseWrite(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	buf := make([]byte, 5000)
	random.NewRand().Read(buf[2500:])

	wrote, err := dagmod.WriteAt(buf[2500:], 2500)
	if err != nil {
		t.Fatal(err)
	}

	if wrote != 2500 {
		t.Fatal("incorrect write amount")
	}

	_, err = dagmod.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(dagmod)
	if err != nil {
		t.Fatal(err)
	}

	if err = testu.ArrComp(out, buf); err != nil {
		t.Fatal(err)
	}
}

func TestSeekPastEndWrite(t *testing.T) {
	runAllSubtests(t, testSeekPastEndWrite)
}

func testSeekPastEndWrite(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	buf := make([]byte, 5000)
	random.NewRand().Read(buf[2500:])

	nseek, err := dagmod.Seek(2500, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	if nseek != 2500 {
		t.Fatal("failed to seek")
	}

	wrote, err := dagmod.Write(buf[2500:])
	if err != nil {
		t.Fatal(err)
	}

	if wrote != 2500 {
		t.Fatal("incorrect write amount")
	}

	_, err = dagmod.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(dagmod)
	if err != nil {
		t.Fatal(err)
	}

	if err = testu.ArrComp(out, buf); err != nil {
		t.Fatal(err)
	}
}

func TestRelativeSeek(t *testing.T) {
	runAllSubtests(t, testRelativeSeek)
}

func testRelativeSeek(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	for i := 0; i < 64; i++ {
		dagmod.Write([]byte{byte(i)})
		if _, err := dagmod.Seek(1, io.SeekCurrent); err != nil {
			t.Fatal(err)
		}
	}

	out, err := io.ReadAll(dagmod)
	if err != nil {
		t.Fatal(err)
	}

	for i, v := range out {
		if v != 0 && i/2 != int(v) {
			t.Errorf("expected %d, at index %d, got %d", i/2, i, v)
		}
	}
}

func TestInvalidSeek(t *testing.T) {
	runAllSubtests(t, testInvalidSeek)
}

func testInvalidSeek(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(t, dserv, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	_, err = dagmod.Seek(10, -10)

	if err != ErrUnrecognizedWhence {
		t.Fatal(err)
	}
}

func TestEndSeek(t *testing.T) {
	runAllSubtests(t, testEndSeek)
}

func testEndSeek(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()

	n := testu.GetEmptyNode(t, dserv, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	_, err = dagmod.Write(make([]byte, 100))
	if err != nil {
		t.Fatal(err)
	}

	offset, err := dagmod.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatal(err)
	}
	if offset != 100 {
		t.Fatal("expected the relative seek 0 to return current location")
	}

	offset, err = dagmod.Seek(0, io.SeekStart)
	if err != nil {
		t.Fatal(err)
	}
	if offset != 0 {
		t.Fatal("expected the absolute seek to set offset at 0")
	}

	offset, err = dagmod.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	if offset != 100 {
		t.Fatal("expected the end seek to set offset at end")
	}
}

func TestReadAndSeek(t *testing.T) {
	runAllSubtests(t, testReadAndSeek)
}

func testReadAndSeek(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()

	n := testu.GetEmptyNode(t, dserv, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	writeBuf := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	dagmod.Write(writeBuf)

	if !dagmod.HasChanges() {
		t.Fatal("there are changes, this should be true")
	}

	readBuf := make([]byte, 4)
	offset, err := dagmod.Seek(0, io.SeekStart)
	if offset != 0 {
		t.Fatal("expected offset to be 0")
	}
	if err != nil {
		t.Fatal(err)
	}

	// read 0,1,2,3
	c, err := dagmod.Read(readBuf)
	if err != nil {
		t.Fatal(err)
	}
	if c != 4 {
		t.Fatalf("expected length of 4 got %d", c)
	}

	for i := byte(0); i < 4; i++ {
		if readBuf[i] != i {
			t.Fatalf("wrong value %d [at index %d]", readBuf[i], i)
		}
	}

	// skip 4
	_, err = dagmod.Seek(1, io.SeekCurrent)
	if err != nil {
		t.Fatalf("error: %s, offset %d, reader offset %d", err, dagmod.curWrOff, getOffset(dagmod.read))
	}

	// read 5,6,7
	readBuf = make([]byte, 3)
	c, err = dagmod.Read(readBuf)
	if err != nil {
		t.Fatal(err)
	}
	if c != 3 {
		t.Fatalf("expected length of 3 got %d", c)
	}

	for i := byte(0); i < 3; i++ {
		if readBuf[i] != i+5 {
			t.Fatalf("wrong value %d [at index %d]", readBuf[i], i)
		}
	}
}

func TestCtxRead(t *testing.T) {
	runAllSubtests(t, testCtxRead)
}

func testCtxRead(t *testing.T, opts testu.NodeOpts) {
	dserv := testu.GetDAGServ()

	n := testu.GetEmptyNode(t, dserv, opts)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		t.Fatal(err)
	}
	if opts.ForceRawLeaves {
		dagmod.RawLeaves = true
	}

	_, err = dagmod.Write([]byte{0, 1, 2, 3, 4, 5, 6, 7})
	if err != nil {
		t.Fatal(err)
	}
	dagmod.Seek(0, io.SeekStart)

	readBuf := make([]byte, 4)
	_, err = dagmod.CtxReadFull(ctx, readBuf)
	if err != nil {
		t.Fatal(err)
	}
	err = testu.ArrComp(readBuf, []byte{0, 1, 2, 3})
	if err != nil {
		t.Fatal(err)
	}
	// TODO(Kubuxu): context cancel case, I will do it after I figure out dagreader tests,
	// because this is exacelly the same.
}

func BenchmarkDagmodWrite(b *testing.B) {
	b.StopTimer()
	dserv := testu.GetDAGServ()
	n := testu.GetEmptyNode(b, dserv, testu.UseProtoBufLeaves)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const wrsize = 4096

	dagmod, err := NewDagModifier(ctx, n, dserv, testu.SizeSplitterGen(512))
	if err != nil {
		b.Fatal(err)
	}

	buf := make([]byte, b.N*wrsize)
	random.NewRand().Read(buf)
	b.StartTimer()
	b.SetBytes(int64(wrsize))
	for i := 0; i < b.N; i++ {
		n, err := dagmod.Write(buf[i*wrsize : (i+1)*wrsize])
		if err != nil {
			b.Fatal(err)
		}
		if n != wrsize {
			b.Fatal("Wrote bad size")
		}
	}
}

func getOffset(reader uio.DagReader) int64 {
	offset, err := reader.Seek(0, io.SeekCurrent)
	if err != nil {
		panic("failed to retrieve offset: " + err.Error())
	}
	return offset
}
