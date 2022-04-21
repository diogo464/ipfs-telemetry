package snapshot

type Sink interface {
	Push(Snapshot)
}
