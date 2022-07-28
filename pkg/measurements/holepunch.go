package measurements

var holepunch HolePunch = nil

type HolePunch interface {
	Register(success bool)
}

func HolePunchRegister(h HolePunch) {
	if holepunch != nil {
		panic("should not happen")
	}
	holepunch = h
}

func WithHolePunch(fn func(HolePunch)) {
	if holepunch != nil {
		fn(holepunch)
	}
}
