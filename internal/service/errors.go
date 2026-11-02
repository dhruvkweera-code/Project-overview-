package service

import "errors"

var ErrNotFound = errors.New("message not found")

// ErrConflict means the message is being processed right now and cannot be changed.
var ErrConflict = errors.New("message is being processed")

// ErrQueueUnavailable means the change was saved but the event could not be sent to Kafka.
var ErrQueueUnavailable = errors.New("message queue unavailable")
