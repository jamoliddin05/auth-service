package closers

import "context"

// Closer is a resource that can be gracefully shut down.
type Closer interface {
	Close(ctx context.Context) error
}
