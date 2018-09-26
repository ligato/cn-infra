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
	"time"

	"github.com/ligato/cn-infra/rpc/rest/security/model/http-security"
)

// StorageType defines user storage type
type StorageType string

// Available storage types
const (
	// DefaultStorageType is type definition of basic AuthStore implementation
	DefaultStorageType StorageType = "default"
)

// AuthStore is common interface to access user database/permissions
type AuthStore interface {
	// AddUser adds new user with name, password and permission groups. Password should be already hashed.
	AddUser(name, passwordHash string, permissions []string) error
	// GetUser returns user data according to name, or nil of not found
	GetUser(name string) (*User, error)
	// SetLoginTime writes last login time for specific user
	SetLoginTime(name string)
	// GetLoginTime returns las login time or error if user does not exist
	GetLoginTime(name string) (time.Time, error)
	// SetLoginTime writes last logout time for specific user
	SetLogoutTime(name string)
	// GetLoginTime returns last logout time or error if user does not exist
	GetLogoutTime(name string) (time.Time, error)
}

// User stores credentials, permissions and tracks last login/logout
type User struct {
	http_security.User
	lastLogin  time.Time
	lastLogout time.Time
}

// defaultAuthStorage is default implementation of AuthStore
type defaultAuthStorage struct {
	db []*User
}

// CreateAuthStore builds new storage type of provided type
func CreateAuthStore(storageType StorageType) AuthStore {
	switch storageType {
	case DefaultStorageType:
		return &defaultAuthStorage{
			db: make([]*User, 0),
		}
	default:
		// Return also as default
		return &defaultAuthStorage{
			db: make([]*User, 0),
		}
	}
}

func (ds *defaultAuthStorage) AddUser(name, passwordHash string, permissions []string) error {
	// Verify user does not exist yet
	_, err := ds.GetUser(name)
	if err == nil {
		// User already exists
		return fmt.Errorf("user %s already exists", name)
	}

	ds.db = append(ds.db, &User{
		User: http_security.User{
			Name:         name,
			PasswordHash: passwordHash,
			Permissions:  permissions,
		},
	})
	return nil
}

func (ds *defaultAuthStorage) GetUser(name string) (*User, error) {
	for _, userData := range ds.db {
		if userData.Name == name {
			return userData, nil
		}
	}
	return nil, fmt.Errorf("user %s not found", name)
}

func (ds *defaultAuthStorage) SetLoginTime(name string) {
	for _, userData := range ds.db {
		if userData.Name == name {
			userData.lastLogin = time.Now()
		}
	}
}

func (ds *defaultAuthStorage) GetLoginTime(name string) (time.Time, error) {
	for _, userData := range ds.db {
		if userData.Name == name {
			return userData.lastLogin, nil
		}
	}
	return time.Time{}, fmt.Errorf("user %s not found", name)
}

func (ds *defaultAuthStorage) SetLogoutTime(name string) {
	for _, userData := range ds.db {
		if userData.Name == name {
			userData.lastLogout = time.Now()
		}
	}
}

func (ds *defaultAuthStorage) GetLogoutTime(name string) (time.Time, error) {
	for _, userData := range ds.db {
		if userData.Name == name {
			return userData.lastLogout, nil
		}
	}
	return time.Time{}, fmt.Errorf("user %s not found", name)
}
