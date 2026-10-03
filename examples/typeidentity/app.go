package typeidentity

type Number = int
type Box[T any] struct{ Value T }

func (b *Box[T]) Close() { var zero T; b.Value = zero }

type Transform = func(Number) struct{ Value Number }
type Options = struct {
	Apply  func(Number) Number
	Values []Box[Number]
}

type Reader = interface {
	Read(Number) Number
	Name() string
}

type memory struct{}

func (memory) Read(n int) int { return n }
func (memory) Name() string   { return "memory" }

func NewTransform() Transform {
	return func(n int) struct{ Value int } { return struct{ Value int }{n + 1} }
}

func NewOptions() Options {
	return Options{Apply: func(n int) int { return n * 2 }, Values: []Box[int]{{Value: 7}}}
}

func NewReader() Reader { return memory{} }

func NewBox() *Box[Number] { return &Box[int]{Value: 3} }

type Result struct {
	Value int
	Box   *Box[int]
}

// Provider signatures deliberately spell identical types differently.
func NewResult(
	f func(int) struct{ Value int },
	o struct {
		Apply  func(int) int
		Values []Box[int]
	},
	r interface {
		Name() string
		Read(int) int
	},
	b *Box[int],
	p struct{ Value int },
	shared Transform,
) Result {
	return Result{Value: f(r.Read(o.Apply(b.Value))).Value + o.Values[0].Value + p.Value + shared(0).Value, Box: b}
}
