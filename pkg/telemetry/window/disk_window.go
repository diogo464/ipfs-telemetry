package window

//const (
//	TelemetryCacheDirectoryName = "ipfs_telemetry"
//	SnapshotsPerDiskBlock       = 1024
//)
//
////var _ Window2 = (*DiskWindow)(nil)
////var _ Fetcher2 = (*diskWindowFetcher)(nil)
//
//type diskWindowBlock struct {
//	session        uuid.UUID
//	first_seqn     uint64
//	last_timestamp time.Time
//	snapshot_count uint64
//	filepath       string
//}
//
//type diskWindowBlockSort []*diskWindowBlock
//
//// Len is the number of elements in the collection.
//func (s diskWindowBlockSort) Len() int {
//	return len(s)
//}
//
//func (s diskWindowBlockSort) Less(i, j int) bool {
//	return s[i].last_timestamp.Before(s[j].last_timestamp)
//}
//
//// Swap swaps the elements with indexes i and j.
//func (s diskWindowBlockSort) Swap(i, j int) {
//	s[i], s[j] = s[j], s[i]
//}
//
//type DiskWindow struct {
//	sync.Mutex
//	session       uuid.UUID
//	duration      time.Duration
//	buf           *vecdeque[windowItem]
//	basedir       string
//	blocks        []*diskWindowBlock
//	pending_flush []*pbs.Snapshot
//}
//
//func NewDiskWindow(duration time.Duration, session uuid.UUID) (*DiskWindow, error) {
//	cachedir, err := os.UserCacheDir()
//	if err != nil {
//		return nil, err
//	}
//	basedir := path.Join(cachedir, TelemetryCacheDirectoryName)
//	blocks, err := readCurrentDiskBlocks(basedir)
//	if err != nil {
//		return nil, err
//	}
//	return &DiskWindow{
//		session:  session,
//		duration: duration,
//		buf:      newVecDeque[windowItem](),
//		basedir:  basedir,
//		blocks:   blocks,
//	}, nil
//}
//
//// Push implements Window2
//func (w *DiskWindow) Push(s snapshot.Snapshot) {
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
//func (w *DiskWindow) NewFetcher(session uuid.UUID, since uint64) Fetcher2 {
//	panic("unimplemented")
//}
//
//func (w *DiskWindow) clean() {
//}
//
//type diskWindowFetcher struct{}
//
//// Close implements Fetcher2
//func (*diskWindowFetcher) Close() {
//	panic("unimplemented")
//}
//
//// Fetch implements Fetcher2
//func (*diskWindowFetcher) Fetch(n int) FetchResult2 {
//	panic("unimplemented")
//}
//
//// Next implements Fetcher2
//func (*diskWindowFetcher) Next() uint64 {
//	panic("unimplemented")
//}
//
//func readCurrentDiskBlocks(dirpath string) ([]*diskWindowBlock, error) {
//	entries, err := ioutil.ReadDir(dirpath)
//	if err != nil {
//		return nil, err
//	}
//
//	blocks := make([]*diskWindowBlock, 0, len(entries))
//	for _, entry := range entries {
//		if !entry.Mode().IsRegular() {
//			continue
//		}
//		filepath := path.Join(dirpath, entry.Name())
//		if block, err := diskWindowBlockReadHeader(filepath); err == nil {
//			blocks = append(blocks, block)
//		}
//	}
//	sort.Sort(diskWindowBlockSort(blocks))
//
//	return blocks, nil
//}
//
//func diskWindowBlockReadHeader(path string) (*diskWindowBlock, error) {
//	f, err := os.Open(path)
//	if err != nil {
//		return nil, err
//	}
//
//	header := new(pbw.WindowHeader)
//	err = pbutils.ReadRle(f, header)
//	if err != nil {
//		return nil, err
//	}
//
//	session, err := uuid.Parse(header.GetSession())
//	if err != nil {
//		return nil, err
//	}
//
//	return &diskWindowBlock{
//		session:        session,
//		first_seqn:     header.GetFirstSeqn(),
//		last_timestamp: header.GetLastTimestamp().AsTime(),
//		snapshot_count: header.GetSnapshotCount(),
//		filepath:       path,
//	}, nil
//}
