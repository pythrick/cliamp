//go:build !unix

package instance

// Lock is a no-op on non-unix platforms.
type Lock struct{}

// Acquire is a no-op on non-unix platforms.
func Acquire(takeover bool) (*Lock, error) { return &Lock{}, nil }

func (l *Lock) Close() {}
