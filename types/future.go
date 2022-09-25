package types

import "fmt"

type future[T any] struct {
	f      func() (T, error)
	result T
	err    error
	stored bool
}

type Future[T any] interface {
	Get() (T, error)
	GetOrPanic() T
	Reset()
	String() string
}

func (f *future[T]) Get() (T, error) {
	if !f.stored && f.f != nil {
		f.result, f.err = f.f()
		f.stored = true
	}

	return f.result, f.err
}

func (f *future[T]) GetOrPanic() T {
	if !f.stored && f.f != nil {
		f.result, f.err = f.f()
		f.stored = true
	}

	if f.err != nil {
		panic(f.err)
	}

	return f.result
}

func (f *future[T]) Reset() {
	f.stored = false
}

func (f *future[T]) String() string {
	return fmt.Sprint(f.GetOrPanic())
}

func FutureFrom[T any](a T) Future[T] {
	return &future[T]{
		f: func() (T, error) {
			return a, nil
		},
	}
}

func FutureFromPtr[T any](a *T) Future[T] {
	return &future[T]{
		f: func() (T, error) {
			return *a, nil
		},
	}
}

func FutureFromFunc[T any](f func() T) Future[T] {
	return &future[T]{
		f: func() (T, error) {
			return f(), nil
		},
	}
}

func FutureFromFuncErr[T any](f func() (T, error)) Future[T] {
	return &future[T]{
		f: f,
	}
}

func FutureInterfacer[T any, I any](f Future[T]) Future[I] {
	return &future[I]{
		f: func() (i I, err error) {
			t, err := f.Get()
			if err != nil {
				return i, err
			}

			i, ok := any(t).(I)
			if !ok {
				return i, fmt.Errorf("cannot convert %T to %T", t, i)
			}

			return i, nil
		},
	}
}

func FutureInterfacerArray[T any, I any](f Future[[]T]) Future[[]I] {
	return &future[[]I]{
		f: func() ([]I, error) {
			t, err := f.Get()
			if err != nil {
				return nil, err
			}

			i := make([]I, len(t))
			var ok bool
			for k, v := range t {
				i[k], ok = any(v).(I)
				if !ok {
					return nil, fmt.Errorf("cannot convert %T to %T", v, i[k])
				}
			}

			return i, nil
		},
	}
}

func FutureFromStringers[T Stringer](f Future[[]T]) Future[[]string] {
	return &future[[]string]{
		f: func() ([]string, error) {
			t, err := f.Get()
			if err != nil {
				return nil, err
			}

			i := make([]string, len(t))
			for k, v := range t {
				i[k] = v.String()
			}

			return i, nil
		},
	}
}
