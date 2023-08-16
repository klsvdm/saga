package saga

import "context"

type Cancellable interface {
	Cancel()
}

type sagaKey struct{}

func FromContext(ctx context.Context) (Cancellable, bool) {
	saga, ok := ctx.Value(sagaKey{}).(Cancellable)
	return saga, ok
}
