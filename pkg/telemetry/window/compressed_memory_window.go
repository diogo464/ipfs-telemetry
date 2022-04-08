package window

//var _ Window = (*CompressedMemoryWindow)(nil)
//
//const CompressedMemoryWindowSnapshotsPerBlock = 256
//
//type compressedMemoryWindowBlock struct {
//	initial_seqn   uint64
//	last_timestamp time.Time
//	data           []byte
//}
//
//func NewCompressedMemoryWindowBlock(in []windowItem) (*compressedMemoryWindowBlock, error) {
//	initial_seqn := in[0].seqn
//	last_timestamp := in[len(in)-1].timestamp
//	var bwriter bytes.Buffer
//	fwriter, err := flate.NewWriter(&bwriter, -1)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, item := range in {
//		marshaled, err := proto.Marshal(item.snapshot)
//		if err != nil {
//			return nil, err
//		}
//		err = rle.Write(fwriter, marshaled)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	err = fwriter.Flush()
//	if err != nil {
//		return nil, err
//	}
//
//	return &compressedMemoryWindowBlock{
//		initial_seqn:   initial_seqn,
//		last_timestamp: last_timestamp,
//		data:           bwriter.Bytes(),
//	}, nil
//}
//
//func (c *compressedMemoryWindowBlock) decompress() ([]*pb.Snapshot, error) {
//	breader := bytes.NewReader(c.data)
//	freader := flate.NewReader(breader)
//	uncompressed, err := io.ReadAll(freader)
//	if err != nil {
//		return nil, err
//	}
//	ureader := bytes.NewReader(uncompressed)
//	snapshots := make([]*pb.Snapshot, 0, CompressedMemoryWindowSnapshotsPerBlock)
//	for {
//		snapshot := new(pb.Snapshot)
//		err = pbutils.ReadRle(ureader, snapshot)
//		if err != nil {
//			break
//		}
//		snapshots = append(snapshots, snapshot)
//	}
//	return snapshots, err
//}
//
//type CompressedMemoryWindow struct {
//	sync.Mutex
//	session  uuid.UUID
//	duration time.Duration
//	buf      *vecdeque[windowItem]
//	blocks   []*compressedMemoryWindowBlock
//}
//
//func NewCompressedMemoryWindow(duration time.Duration, session uuid.UUID) *CompressedMemoryWindow {
//	return &CompressedMemoryWindow{
//		session:  session,
//		duration: duration,
//		buf:      newVecDeque[windowItem](),
//		blocks:   []*compressedMemoryWindowBlock{},
//	}
//}
//
//// Push implements Window2
//func (w *CompressedMemoryWindow) Push(s snapshot.Snapshot) {
//	w.Lock()
//	defer w.Unlock()
//	w.buf.PushBack(windowItem{
//		seqn:      nextSeqN(w.buf),
//		snapshot:  s.ToPB(),
//		timestamp: s.GetTimestamp(),
//	})
//	w.clean()
//}
//
//// NewFetcher implements Window2
//func (w *CompressedMemoryWindow) NewFetcher(session uuid.UUID, since uint64) Fetcher2 {
//	if session != w.session {
//		since = 0
//	}
//
//	return &compressedMemoryWindowFetcher{
//		window: w,
//		since:  since,
//		done:   false,
//	}
//}
//
//func (w *CompressedMemoryWindow) clean() {
//	if w.buf.Len() >= CompressedMemoryWindowSnapshotsPerBlock {
//		buf := w.buf.TakeBuffer()
//		block, err := NewCompressedMemoryWindowBlock(buf)
//		if err != nil {
//			panic(err)
//		}
//		w.blocks = append(w.blocks, block)
//	}
//}
//
//type compressedMemoryWindowFetcher struct {
//	window *CompressedMemoryWindow
//	since  uint64
//	done   bool
//}
//
//// Close implements Fetcher2
//func (f *compressedMemoryWindowFetcher) Close() {}
//
//// Fetch implements Fetcher2
//func (f *compressedMemoryWindowFetcher) Fetch(n int) FetchResult2 {
//	if f.done {
//		return FetchResult2{Err: io.EOF}
//	}
//
//	f.window.Lock()
//	defer f.window.Unlock()
//	if f.window.buf.IsEmpty() {
//		f.done = true
//		return FetchResult2{Err: io.EOF}
//	}
//
//	start := sort.Search(f.window.buf.Len(), func(i int) bool {
//		return f.window.buf.Get(i).seqn >= f.since
//	})
//	end := utils.Min(start+n, f.window.buf.Len())
//	snapshots := make([]*pb.Snapshot, 0, end-start)
//
//	for i := start; i < end; i++ {
//		snapshots = append(snapshots, f.window.buf.Get(i).snapshot)
//	}
//
//	if end < f.window.buf.Len() {
//		f.since = f.window.buf.Get(end).seqn
//	} else {
//		f.done = true
//	}
//
//	return FetchResult2{
//		Session:   f.window.session,
//		Snapshots: snapshots,
//		Err:       nil,
//	}
//}
//
//func (f *compressedMemoryWindowFetcher) Next() uint64 {
//	return f.since
//}
