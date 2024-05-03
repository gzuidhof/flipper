package resource

import "errors"

// ErrNilResource is returned when a nil resource is unexpectedly used.
var ErrNilResource = errors.New("resource is nil")

// ErrWrongProvider is returned when a resource is used with the wrong provider.
var ErrWrongProvider = errors.New("resource is from the wrong provider")
