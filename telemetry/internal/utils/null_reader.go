package utils

type NullReader struct{}

func (NullReader) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = 0
	}
	n = len(p)
	err = nil
	return
}
