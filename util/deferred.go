package util

import "sync"

type Deferred[T any, E any] struct {
	fulfilled chan struct{}
	once      sync.Once
	val       *T
	err       *E
}

func MakeDeferred[T any, E any]() *Deferred[T, E] {
	deferred := Deferred[T, E]{
		fulfilled: make(chan struct{}),
		val:       nil,
		err:       nil,
	}
	return &deferred
}

func (d *Deferred[T, E]) Resolve(val T) {
	d.once.Do(func() {
		d.val = &val
		close(d.fulfilled)
	})
}

func (d *Deferred[T, E]) Reject(err E) {
	d.once.Do(func() {
		d.err = &err
		close(d.fulfilled)
	})
}

func (d *Deferred[T, E]) Wait() (*T, *E) {
	<-d.fulfilled
	return d.val, d.err
}
