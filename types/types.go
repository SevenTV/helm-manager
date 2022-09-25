package types

type Stringer interface {
	String() string
}

type StringerFunc func() string

func (s StringerFunc) String() string {
	return s()
}

type ToStringer string

func (s ToStringer) String() string {
	return string(s)
}
