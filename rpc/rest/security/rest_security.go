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

//go:generate protoc --proto_path=model/http-security --gogo_out=model/http-security model/http_security/httpsecurity.proto

package security

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/ligato/cn-infra/logging"
	httpsecurity "github.com/ligato/cn-infra/rpc/rest/security/model/http-security"
	"github.com/pkg/errors"
	"github.com/unrolled/render"
	"golang.org/x/crypto/bcrypt"
)

const (
	// Helps to obtain authorization header matching the field in a request
	authHeaderStr = "authorization"
	// Used to sign the token // todo maybe use something more secure
	secret = "secret"
	// Admin constant, used to define admin security group and user
	admin = "admin"
)

const (
	// URL for login. Successful login returns token. Re-login invalidates old token and returns a new one.
	login = "/login"
	// URL key for logout, invalidates current token.
	logout = "/logout"
)

// AuthenticatorAPI provides methods for handling permissions
type AuthenticatorAPI interface {
	// AddPermissionGroup adds new permission group. PG is defined by name and a set of URL keys. User with
	// permission group enabled has access to that set of keys. PGs with duplicated names are skipped.
	AddPermissionGroup(group ...*httpsecurity.PermissionGroup)

	// Validate serves as middleware used while registering new HTTP handler. For every request, token
	// and permission group is validated.
	Validate(provider http.HandlerFunc) http.HandlerFunc
}

// Authenticator keeps information about users, permission groups and tokens and processes it
type authenticator struct {
	log logging.Logger

	// Router instance automatically registers login/logout REST API handlers if authentication is enabled
	router    *mux.Router
	formatter *render.Render

	// User database keeps all known users with permissions and hashed password. Users are loaded from
	// HTTP config file // todo add other options
	userDb []*httpsecurity.User
	// Permission database is a map of permissions and bound URLs
	groupDb []*httpsecurity.PermissionGroup
	// Token database keeps information of actual token and its owner.
	tokenDb map[string]string

	// Token claims
	expTime int
}

// NewAuthenticator prepares new instance of authenticator.
func NewAuthenticator(router *mux.Router, expTime int, users []httpsecurity.User, cost int, log logging.Logger) AuthenticatorAPI {
	a := &authenticator{
		router: router,
		log:    log,
		formatter: render.New(render.Options{
			IndentJSON: true,
		}),
		userDb:  make([]*httpsecurity.User, 0),
		tokenDb: make(map[string]string),
		expTime: expTime,
	}

	// Add admin-user, enabled by default, always has access to every URL
	users = append(users, httpsecurity.User{
		Name:        admin,
		Password:    "ligato123",
		Permissions: []string{admin},
	})

	go func() {
		// Process users in go routine, since hashing may take some time
		for _, user := range users {
			// Password hash cost is currently set via config file. Values allowed by bcrypt are in range 4-31. The min
			// value will be set if cost is not defined in config file. Keep in mind that high cost values (>10) can
			// take a lot of time to process.
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), cost)
			if err != nil {
				a.log.Errorf("failed to hash password for user %s: %v", user.Name, err)
				continue
			}

			a.userDb = append(a.userDb, &httpsecurity.User{
				Name:        user.Name,
				Password:    string(hash),
				Permissions: user.Permissions,
			})
			a.log.Debugf("Registered user %s, permissions: %v", user.Name, user.Permissions)
		}
	}()

	// Admin-group, available by default and always enabled for all URLs
	a.AddPermissionGroup(&httpsecurity.PermissionGroup{
		Name: admin,
	})

	a.registerSecurityHandlers()

	return a
}

// AddPermissionGroup adds new permission group.
func (a *authenticator) AddPermissionGroup(group ...*httpsecurity.PermissionGroup) {
	for _, newPermissionGroup := range group {
		for _, existingGroup := range a.groupDb {
			if existingGroup.Name == newPermissionGroup.Name {
				a.log.Warnf("permission group %s already exists, skipped")
				continue
			}
		}
		a.log.Debugf("added HTTP permission group %s", newPermissionGroup.Name)
		a.groupDb = append(a.groupDb, newPermissionGroup)
	}
}

// Validate the request
func (a *authenticator) Validate(provider http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authHeader := req.Header.Get(authHeaderStr)
		if authHeader == "" {
			a.formatter.Text(w, http.StatusUnauthorized, "401 Unauthorized: authorization header required")
			return
		}
		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			a.formatter.Text(w, http.StatusUnauthorized, "401 Unauthorized: invalid authorization token")
			return
		}
		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("error parsing token")
			}
			return []byte(secret), nil
		})
		if err != nil {
			errStr := fmt.Sprintf("500 internal server error: %s", err)
			a.formatter.Text(w, http.StatusInternalServerError, errStr)
			return
		}

		if err := a.validateToken(token.Raw, req.URL.Path); err != nil {
			errStr := fmt.Sprintf("401 Unauthorized: %v", err)
			a.formatter.Text(w, http.StatusUnauthorized, errStr)
			return
		}

		provider.ServeHTTP(w, req)
	})
}

// Register authenticator-wide security handlers
func (a *authenticator) registerSecurityHandlers() {
	a.router.HandleFunc(login, a.createTokenEndpoint).Methods(http.MethodPost)
	a.router.HandleFunc(logout, a.invalidateTokenEndpoint).Methods(http.MethodPost)
}

// Validates credentials and provides new token
func (a *authenticator) createTokenEndpoint(w http.ResponseWriter, req *http.Request) {
	name, errCode, err := a.validateCredentials(req)
	if err != nil {
		a.formatter.Text(w, errCode, err.Error())
		return
	}
	var claims jwt.MapClaims = make(map[string]interface{})
	if a.expTime != 0 {
		claims["exp"] = time.Now().Add(time.Minute * time.Duration(a.expTime)).Unix()
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		errStr := fmt.Sprintf("500 internal server error: faield to sign token: %v", err)
		a.formatter.Text(w, http.StatusInternalServerError, errStr)
		return
	}
	a.tokenDb[name] = tokenString
	a.formatter.Text(w, http.StatusOK, tokenString)
	return
}

// Removes token endpoint from the DB. During processing, token will not be found and will be considered as invalid.
func (a *authenticator) invalidateTokenEndpoint(w http.ResponseWriter, req *http.Request) {
	name, _, ok := req.BasicAuth()
	if !ok {
		a.formatter.Text(w, http.StatusInternalServerError, "500 internal server error: unable to process authentication data")
		return
	}
	delete(a.tokenDb, name)
}

// Validates credentials, returns name and error code/message if invalid
func (a *authenticator) validateCredentials(req *http.Request) (string, int, error) {
	name, pass, ok := req.BasicAuth()
	if !ok {
		return name, http.StatusInternalServerError, errors.Errorf("500 internal server error: unable to process authentication data")
	}
	user := a.getUser(name)
	if user == nil {
		return name, http.StatusUnauthorized, errors.Errorf("401 unauthorized: user name does not exist")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass)); err != nil {
		return name, http.StatusUnauthorized, fmt.Errorf("401 unauthorized: incorrect password")
	}
	return name, 0, nil
}

// Validates token itself and permissions
func (a *authenticator) validateToken(token, url string) error {
	owner, err := a.getTokenOwner(token)
	if err != nil {
		return err
	}
	user := a.getUser(owner)
	if user == nil {
		return fmt.Errorf("user name does not exist")
	}
	// Do not check for permissions if user is admin
	if userIsAdmin(user) {
		return nil
	}

	perms := a.getPermissionsForURL(url)
	for _, userPerm := range user.Permissions {
		for _, perm := range perms {
			if userPerm == perm {
				return nil
			}
		}
	}

	return fmt.Errorf("not permitted")
}

// Returns user data according to name, nil if does not exists
func (a *authenticator) getUser(name string) *httpsecurity.User {
	for _, userData := range a.userDb {
		if userData.Name == name {
			return userData
		}
	}
	return nil
}

// Returns token owner, or error if not found
func (a *authenticator) getTokenOwner(token string) (string, error) {
	for name, knownToken := range a.tokenDb {
		if token == knownToken {
			return name, nil
		}
	}
	return "", fmt.Errorf("authorization token is invalid")
}

// Returns all permission groups provided URL is allowed for
func (a *authenticator) getPermissionsForURL(url string) []string {
	var groups []string
	for _, group := range a.groupDb {
		for _, groupURL := range group.Urls {
			if groupURL == url {
				groups = append(groups, group.Name)
			}
		}
	}
	return groups
}

// Checks user admin permission
func userIsAdmin(user *httpsecurity.User) bool {
	for _, permission := range user.Permissions {
		if permission == admin {
			return true
		}
	}
	return false
}
