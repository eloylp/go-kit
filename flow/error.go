package flow

import (
	"errors"
)

var (
	ErrSubscriberNotFound = errors.New("fanout: subscriber not found")
)
