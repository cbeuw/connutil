package connutil

import "errors"

var (
	ErrTimeout = errors.New("deadline exceeded")
)
