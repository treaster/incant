package interpreter

type Context map[string]int

type Expression interface {
	Eval(Context, *int) error
	String() string
}

type Statement interface {
	Exec(Context) error
	String() string
}
