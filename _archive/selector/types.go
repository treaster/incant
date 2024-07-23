package selector

type Context map[string]any

type Expression interface {
	Eval(Context, *int) error
	String() string
}
