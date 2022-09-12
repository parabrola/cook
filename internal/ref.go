package internal

type Ref[T any] struct {
	value T
	err   error
}

func NewRef[T any](value T, err error) Ref[T] {
	return Ref[T]{
		value: value,
		err:   err,
	}
}

func (r *Ref[T]) Value() T {
	return r.value
}

func (r *Ref[T]) Error() error {
	return r.err
}
