//  Copyright (c) 2018 Cisco and/or its affiliates.
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
	"context"
	"fmt"
	"net/http"
	"strings"
)

type contextKey int

const (
	userKey contextKey = iota
)

// UserName returns name of authorized user for the request.
func UserName(r *http.Request) string {
	if val := r.Context().Value(userKey); val != nil {
		return val.(string)
	}
	return ""
}

func setUserName(r *http.Request, username string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), userKey, username))
}

// staticAuthenticator is default implementation of BasicHTTPAuthenticator
type staticAuthenticator struct {
	credentials map[string]string
}

// newStaticAuthenticator creates new instance of static authenticator.
// Argument `users` is a slice of colon-separated username and password couples.
func newStaticAuthenticator(users []string) (*staticAuthenticator, error) {
	sa := &staticAuthenticator{
		credentials: make(map[string]string),
	}
	for _, u := range users {
		fields := strings.Split(u, ":")
		if len(fields) != 2 {
			return nil, fmt.Errorf("invalid format of basic auth entry '%v' expected 'user:pass'", u)
		}
		sa.credentials[fields[0]] = fields[1]
	}
	return sa, nil
}

// Authenticate looks up the given user name and password in the internal map.
// If match is found returns true, false otherwise.
func (sa *staticAuthenticator) Authenticate(user string, pass string) bool {
	password, found := sa.credentials[user]
	if !found {
		return false
	}
	return pass == password
}
