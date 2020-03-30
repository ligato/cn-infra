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

package security

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	access "go.ligato.io/cn-infra/v2/rpc/rest/security/model/access-security"
)

// AuthenticationDB is common interface to access user database/permissions
type AuthenticationDB interface {
	// AddUser adds new user with name, password and permission groups. Password should be already hashed.
	AddUser(name, password string, permissions []string) error
	// GetUser returns user data according to name, or nil of not found
	GetUser(name string) (*User, error)
	// Authenticate authenticates user with password.
	Authenticate(name, password string) error
	// SetLoginTime writes last login time for specific user
	SetLoginTime(name string)
	// SetLogoutTime writes last logout time for specific user
	SetLogoutTime(name string)
	// IsLoggedOut uses login/logout timestamps to evaluate whether the user was logged out
	IsLoggedOut(name string) (bool, error)
}

// User stores credentials, permissions and tracks last login/logout
type User struct {
	access.User
	lastLogin  time.Time
	lastLogout time.Time
}

// defaultAuthStorage is default implementation of AuthStore
type memoryAuthDB struct {
	sync.Mutex
	cost int
	db   []*User
}

// CreateDefaultAuthDB builds new default storage
func CreateDefaultAuthDB(cost int) AuthenticationDB {
	return &memoryAuthDB{
		cost: cost,
		db:   make([]*User, 0),
	}
}

func (ds *memoryAuthDB) AddUser(name, password string, permissions []string) error {
	ds.Lock()
	defer ds.Unlock()

	// Verify user does not exist yet
	for _, userData := range ds.db {
		if userData.Name == name {
			return fmt.Errorf("user %s already exists", name)
		}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), ds.cost)
	if err != nil {
		return err
	}

	ds.db = append(ds.db, &User{
		User: access.User{
			Name:        name,
			Password:    string(passwordHash),
			Permissions: permissions,
		},
	})
	return nil
}

func (ds *memoryAuthDB) GetUser(name string) (*User, error) {
	ds.Lock()
	defer ds.Unlock()

	for _, userData := range ds.db {
		if userData.Name == name {
			return userData, nil
		}
	}
	return nil, fmt.Errorf("user %s not found", name)
}

func (ds *memoryAuthDB) Authenticate(name, password string) error {
	user, err := ds.GetUser(name)
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return ErrInvalidUsernameOrPassword
	}
	return nil
}

func (ds *memoryAuthDB) SetLoginTime(name string) {
	ds.Lock()
	defer ds.Unlock()

	for _, userData := range ds.db {
		if userData.Name == name {
			userData.lastLogin = time.Now()
		}
	}
}

func (ds *memoryAuthDB) SetLogoutTime(name string) {
	ds.Lock()
	defer ds.Unlock()

	for _, userData := range ds.db {
		if userData.Name == name {
			userData.lastLogout = time.Now()
		}
	}
}

func (ds *memoryAuthDB) IsLoggedOut(name string) (bool, error) {
	ds.Lock()
	defer ds.Unlock()

	for _, user := range ds.db {
		if user.Name == name {
			if user.lastLogout.After(user.lastLogin) {
				return true, nil
			}
			return false, nil
		}
	}
	return false, fmt.Errorf("user %s not found", name)
}
