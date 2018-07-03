package api

import (
	"errors"
)

var (
	// ErrDumpNotSupported should be returned by Dump when dumping is not supported.
	ErrDumpNotSupported = errors.New("dump operation is not supported")

	// ErrCombinedResyncWithChange is returned when transaction combines resync with data changes.
	ErrCombinedResyncWithChange = errors.New("resync combined with data changes in one transaction")

	// ErrClosedScheduler is returned when scheduler is closed during transaction execution.
	ErrClosedScheduler = errors.New("scheduler was closed")

	// ErrTxnWaitCanceled is returned when waiting for result of blocking transaction is canceled.
	ErrTxnWaitCanceled = errors.New("waiting for result of blocking transaction was canceled")

	// ErrTxnQueueFull is returned when the queue of pending transactions is full.
	ErrTxnQueueFull = errors.New("transaction queue is full")

	// ErrUnimplementedKey is returned for Object or Action values without provided descriptor.
	ErrUnimplementedKey = errors.New("unimplemented key")
)
