// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
