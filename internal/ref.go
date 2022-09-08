package internal

type Ref[T comparable] struct {
	value T
	err   error
}

func NewRef[T comparable](value T, err error) Ref[T] {
	return Ref[T]{
		value: value,
		err:   err,
	}
}

func (r *Ref[T]) Equal(value T) bool {
	return r.value == value
}

func (r *Ref[T]) Value() T {
	return r.value
}

func (r *Ref[T]) Error() error {
	return r.err
}
