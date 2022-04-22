package datapoint

type Sink interface {
	Push(Datapoint)
}
