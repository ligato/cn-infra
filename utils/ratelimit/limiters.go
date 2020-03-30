//  Copyright (c) 2020 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package ratelimit

import (
	"sync"

	"golang.org/x/time/rate"
)

// Limiters provides map of rate limiters per X (user, IP..).
type Limiters struct {
	limiters *sync.Map
	rate     rate.Limit
	burst    int
}

func NewLimiter(r rate.Limit, burst int) *Limiters {
	i := &Limiters{
		limiters: new(sync.Map),
		rate:     r,
		burst:    burst,
	}
	return i
}

// Add creates a new rate limiter and adds it to the map.
func (l *Limiters) Add(key string) *rate.Limiter {
	limiter := rate.NewLimiter(l.rate, l.burst)

	l.limiters.Store(key, limiter)

	return limiter
}

// Get returns the rate limiter for the provided key if it exists,
// otherwise calls Add to add key to the map.
func (l *Limiters) Get(key string) *rate.Limiter {
	if limiter, ok := l.limiters.Load(key); ok {
		return limiter.(*rate.Limiter)
	}

	return l.Add(key)
}

func (l *Limiters) Allow(key string) bool {
	limiter := l.Get(key)
	return limiter.Allow()
}
