package flow

import (
	"errors"
)

var (
	// ErrSubscriberNotFound error will be propagated when no subscriber
	// is found.
	ErrSubscriberNotFound = errors.New("fanout: subscriber not found")
)
