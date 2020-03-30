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

package rest

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"go.ligato.io/cn-infra/v2/utils/ratelimit"
)

const (
	HeaderKeyAuthUsername = "X-Ligato-Auth-Username"

	HeaderKeyRateLimitLimit = "X-Ligato-RateLimit-Limit"
	HeaderKeyRateLimitBurst = "X-Ligato-RateLimit-Burst"
)

func (p *Plugin) permMiddleware(isPermitted func(user string, r *http.Request) error) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userName := UserName(r)
			if err := isPermitted(userName, r); err != nil {
				p.Log.Warnf("permission denied for user: %v, path: %v", userName, r.URL.Path)
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (p *Plugin) authMiddleware(authorize func(r *http.Request) (string, error)) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userName, err := authorize(r)
			if err != nil {
				p.Log.Warnf("authorization failed for user: %v", userName)
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			if userName != "" {
				w.Header().Set(HeaderKeyAuthUsername, userName)
			}

			next.ServeHTTP(w, setUserName(r, userName))
		})
	}
}

func rateLimitMiddleware(limiters *ratelimit.Limiters) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)

			var limitBy = ip
			if userName := UserName(r); userName != "" {
				limitBy = userName
			}

			limiter := limiters.Get(limitBy)
			w.Header().Set(HeaderKeyRateLimitLimit, fmt.Sprint(limiter.Limit()))
			w.Header().Set(HeaderKeyRateLimitBurst, fmt.Sprint(limiter.Burst()))

			if !limiter.Allow() {
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
