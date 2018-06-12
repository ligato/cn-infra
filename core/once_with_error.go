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

package core

import "sync"

// OnceWithError is a wrapper around sync.Once that properly handles:
// func() error
// instead of just
// func()
type OnceWithError struct {
	once sync.Once
	err  error
}

// Do provides the same functionality as sync.Once.Do(func()) but for
// func() error
func (owe *OnceWithError) Do(f func() error) error {
	owe.once.Do(func() {
		owe.err = f()
	})
	return owe.err
}