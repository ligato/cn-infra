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

//go:generate protoc --proto_path=model/access-security --go_out=model/access-security model/access-security/accesssecurity.proto

package security

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/unrolled/render"

	"go.ligato.io/cn-infra/v2/logging"
	access "go.ligato.io/cn-infra/v2/rpc/rest/security/model/access-security"
)

var (
	ErrPermissionDenied          = errors.New("auth: permission denied")
	ErrInvalidAuthToken          = errors.New("auth: invalid auth token")
	ErrInvalidUsernameOrPassword = errors.New("auth: invalid username or password")
)

const (
	// Returns login page where credentials may be put. Redirects to authenticate, and if successful, moves to index.
	login = "/login"
	// URL key for logout, invalidates current token.
	logout = "/logout"
	// Authentication page, validates credentials and if successful, returns a token or writes a cookie to a browser
	authenticate = "/authenticate"
	// Cookie name identifier
	cookieName = "ligato-rest-auth"
	// AuthHeaderKey helps to obtain authorization header matching the field in a request
	AuthHeaderKey = "authorization"
)

// Default values, if not provided from config file
const (
	defaultUser    = "admin"
	defaultPass    = "ligato123"
	defaultSignKey = "secret"
	defaultExpTime = time.Hour
)

// AuthenticatorAPI provides methods for handling permissions
type AuthenticatorAPI interface {
	// RegisterHandlers registers authenticator handlers to router.
	RegisterHandlers(router *mux.Router)

	// AddPermissionGroup adds new permission group. PG is defined by name and
	// a set of URL keys. User with permission group enabled has access to that
	// set of keys. PGs with duplicated names are skipped.
	AddPermissionGroup(group ...*access.PermissionGroup)

	// Validate provides middleware used while registering new HTTP handler.
	// For every request, token and permission group is validated.
	Validate(h http.Handler) http.Handler

	// AuthorizeRequest tries to authorize user from request.
	AuthorizeRequest(r *http.Request) (user string, err error)

	// IsPermitted checks if user is permitted to access URL from request.
	IsPermitted(user string, r *http.Request) error
}

// Settings defines fields required to instantiate authenticator
type Settings struct {
	// Router
	Router *mux.Router
	// Authentication database, default implementation is used if not set
	AuthStore AuthenticationDB
	// List of registered users
	Users []access.User
	// Expiration time (token claim). If not set, default value of 1 hour will be used.
	ExpTime time.Duration
	// Cost value used to hash user passwords
	Cost int
	// Custom token sign key. If not set, default value will be used.
	SignKey string
}

// Credentials struct represents simple user login input
type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Authenticator keeps information about users, permission groups and tokens and processes it
type authenticator struct {
	log logging.Logger

	urlPaths  []string
	formatter *render.Render

	authDB  AuthenticationDB
	groupDb map[string][]*access.PermissionGroup_Permissions

	expTime time.Duration
	signKey string
}

// NewAuthenticator prepares new instance of authenticator.
func NewAuthenticator(opt *Settings, log logging.Logger) AuthenticatorAPI {
	a := &authenticator{
		log: log,
		formatter: render.New(render.Options{
			IndentJSON: true,
		}),
		groupDb: make(map[string][]*access.PermissionGroup_Permissions),
		expTime: defaultExpTime,
		signKey: defaultSignKey,
	}

	// Authentication store
	if opt.AuthStore != nil {
		a.authDB = opt.AuthStore
	} else {
		a.authDB = CreateDefaultAuthDB(opt.Cost)
	}

	if opt.SignKey != "" {
		a.signKey = opt.SignKey
	}
	if opt.ExpTime != 0 {
		a.expTime = opt.ExpTime
	}
	a.log.Debugf("Token expiration time set to %v", a.expTime)

	// Hash of default admin password, hashed with cost 10
	if err := a.authDB.AddUser(defaultUser, defaultPass, []string{defaultUser}); err != nil {
		a.log.Errorf("failed to add admin user: %v", err)
	}

	for _, user := range opt.Users {
		if user.Name == defaultUser {
			a.log.Errorf("rejected to create user-defined account named 'admin'")
			continue
		}
		if err := a.authDB.AddUser(user.Name, user.Password, user.Permissions); err != nil {
			a.log.Errorf("failed to add user %s: %v", user.Name, err)
			continue
		}
		a.log.Debug("Registered user %s, permissions: %v", user.Name, user.Permissions)
	}

	// Admin-group, available by default and always enabled for all URLs
	a.groupDb[defaultUser] = []*access.PermissionGroup_Permissions{}

	return a
}

// RegisterHandlers authenticator-wide security handlers
func (a *authenticator) RegisterHandlers(router *mux.Router) {
	add := func(r *mux.Route) {
		path, err := r.URLPath()
		if err != nil {
			panic(err)
		}
		a.urlPaths = append(a.urlPaths, path.Path)
	}

	add(router.HandleFunc(login, a.loginHandler).Methods(http.MethodGet, http.MethodPost))
	add(router.HandleFunc(authenticate, a.authenticationHandler).Methods(http.MethodPost))
	add(router.HandleFunc(logout, a.logoutHandler).Methods(http.MethodPost))
}

func (a *authenticator) requiresAuth(req *http.Request) bool {
	for _, path := range a.urlPaths {
		if req.URL.Path == path {
			return false
		}
	}
	return true
}

// AddPermissionGroup adds new permission group.
func (a *authenticator) AddPermissionGroup(group ...*access.PermissionGroup) {
	for _, newPermissionGroup := range group {
		if _, ok := a.groupDb[newPermissionGroup.Name]; ok {
			a.log.Warnf("permission group %s already exists", newPermissionGroup.Name)
			continue
		}
		a.groupDb[newPermissionGroup.Name] = newPermissionGroup.Permissions
		a.log.Debugf("added permission group %s", newPermissionGroup.Name)
	}
}

func (a *authenticator) AuthorizeRequest(req *http.Request) (string, error) {
	if !a.requiresAuth(req) {
		return "", nil
	}

	// Token may be accessed via cookie, or from authentication header
	tokenString, err := a.getTokenStringFromRequest(req)
	if err != nil {
		return "", err
	}
	// Retrieve token object from raw string
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := jwt.GetSigningMethod(token.Header["alg"].(string)).(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidAuthToken
		}
		return []byte(a.signKey), nil
	})
	if err != nil {
		return "", err
	}
	// Validate token claims
	if token.Claims != nil {
		if err := token.Claims.Valid(); err != nil {
			return "", err
		}
	}
	// Authorize user
	user, err := a.authorizeUser(token)
	if err != nil {
		return "", err
	}

	return user.Name, nil
}

func (a *authenticator) IsPermitted(userName string, req *http.Request) error {
	if !a.requiresAuth(req) {
		return nil
	}

	user, err := a.authDB.GetUser(userName)
	if err != nil {
		return fmt.Errorf("invalid user: %v", err)
	}
	if err := a.checkUserPermission(user, req.URL.Path, req.Method); err != nil {
		return err
	}
	return nil
}

// Validate the request
func (a *authenticator) Validate(provider http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Token may be accessed via cookie, or from authentication header
		tokenString, err := a.getTokenStringFromRequest(req)
		if err != nil {
			a.formatter.Text(w, http.StatusUnauthorized, err.Error())
			return
		}
		// Retrieve token object from raw string
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := jwt.GetSigningMethod(token.Header["alg"].(string)).(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidAuthToken
			}
			return []byte(a.signKey), nil
		})
		if err != nil {
			a.formatter.Text(w, http.StatusUnauthorized, err.Error())
			return
		}
		// Validate token claims
		if token.Claims != nil {
			if err := token.Claims.Valid(); err != nil {
				a.formatter.Text(w, http.StatusUnauthorized, err.Error())
				return
			}
		}

		user, err := a.authorizeUser(token)
		if err != nil {
			a.formatter.Text(w, http.StatusUnauthorized, err.Error())
			return
		}
		err = a.checkUserPermission(user, req.URL.Path, req.Method)
		if err != nil {
			a.formatter.Text(w, http.StatusForbidden, err.Error())
			return
		}

		provider.ServeHTTP(w, req)
	})
}

// Login handler shows simple page to log in
func (a *authenticator) loginHandler(w http.ResponseWriter, req *http.Request) {
	// GET returns login page. Submit redirects to authenticate.
	if req.Method == http.MethodGet {
		if err := loginFormTmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if req.Method == http.MethodPost {
		// POST decodes provided credentials
		credentials := &credentials{}
		err := json.NewDecoder(req.Body).Decode(&credentials)
		if err != nil {
			errStr := fmt.Sprintf("decoding credentials failed: %v", err)
			a.formatter.Text(w, http.StatusBadRequest, errStr)
			return
		}
		token, errCode, err := a.getTokenFor(credentials)
		if err != nil {
			a.formatter.Text(w, errCode, err.Error())
			return
		}

		// Returns token string.
		a.formatter.Text(w, http.StatusOK, token)
	} else {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// Authentication handler verifies credentials from login page (GET) and writes cookie with token
func (a *authenticator) authenticationHandler(w http.ResponseWriter, req *http.Request) {
	// Read name and password from the form (if accessed from browser)
	credentials := &credentials{
		Username: req.FormValue("username"),
		Password: req.FormValue("password"),
	}
	token, errCode, err := a.getTokenFor(credentials)
	if err != nil {
		a.formatter.Text(w, errCode, err.Error())
		return
	}

	// Writes cookie with token.
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Path:   "/",
		MaxAge: int(a.expTime.Seconds()),
		Value:  token,
		Secure: false,
	})
	// Automatically move to index page.
	http.Redirect(w, req, "/", http.StatusMovedPermanently)
}

// Removes token endpoint from the DB. During processing, token will not be found and will be considered as invalid.
func (a *authenticator) logoutHandler(w http.ResponseWriter, req *http.Request) {
	var credentials credentials
	err := json.NewDecoder(req.Body).Decode(&credentials)
	if err != nil {
		errStr := fmt.Sprintf("failed to decode json: %v", err)
		a.formatter.Text(w, http.StatusBadRequest, errStr)
		return
	}

	a.authDB.SetLogoutTime(credentials.Username)
	a.log.Debugf("user %s has logged out", credentials.Username)
}

// Read raw token from request.
func (a *authenticator) getTokenStringFromRequest(req *http.Request) (result string, err error) {
	// Try to read header, validate it if exists.
	authHeader := req.Header.Get(AuthHeaderKey)
	if authHeader != "" {
		return parseTokenFromAuthHeader(authHeader)
	}
	a.log.Debugf("authorization header not found, checking cookies..")

	// Otherwise read cookie
	cookie, err := req.Cookie(cookieName)
	if err == nil && cookie != nil {
		return cookie.Value, nil
	}
	a.log.Debugf("authorization cookie not found (err: %v)", err)

	return "", fmt.Errorf("authorization required")
}

func parseTokenFromAuthHeader(authHeader string) (string, error) {
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 {
		return "", ErrInvalidAuthToken
	}
	// Parse token header constant
	if parts[0] != "Bearer" {
		return "", ErrInvalidAuthToken
	}
	return parts[1], nil
}

// Get token for credentials
func (a *authenticator) getTokenFor(cred *credentials) (string, int, error) {
	err := a.authDB.Authenticate(cred.Username, cred.Password)
	if err != nil {
		a.log.Warnf("authentication failed for user: %v", cred.Username)
		return "", http.StatusUnauthorized, err
	}

	claims := jwt.StandardClaims{
		Audience:  cred.Username,
		ExpiresAt: a.expTime.Nanoseconds(),
		Issuer:    "cn-infra.ligato.io",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.signKey))
	if err != nil {
		return "", http.StatusInternalServerError, fmt.Errorf("failed to sign token: %v", err)
	}

	a.authDB.SetLoginTime(cred.Username)
	a.log.Infof("user %s has logged in", cred.Username)

	return tokenString, 0, nil
}

func (a *authenticator) userNameFromToken(token *jwt.Token) (string, error) {
	var userName string
	// Read audience from the token
	switch v := token.Claims.(type) {
	case jwt.MapClaims:
		var ok bool
		if userName, ok = v["aud"].(string); !ok {
			return "", fmt.Errorf("failed to validate token claims audience")
		}
	case jwt.StandardClaims:
		userName = v.Audience
	default:
		return "", fmt.Errorf("failed to validate token claims")
	}
	return userName, nil
}

func (a *authenticator) authorizeUser(token *jwt.Token) (*User, error) {
	userName, err := a.userNameFromToken(token)
	if err != nil {
		return nil, err
	}
	loggedOut, err := a.authDB.IsLoggedOut(userName)
	if err != nil {
		return nil, fmt.Errorf("authorizing user failed: %v", err)
	}
	// User logged out
	if loggedOut {
		token.Valid = false
		return nil, fmt.Errorf("user has logged out")
	}
	user, err := a.authDB.GetUser(userName)
	if err != nil {
		return nil, fmt.Errorf("invalid user: %v", err)
	}
	return user, nil
}

func (a *authenticator) checkUserPermission(user *User, url, method string) error {
	// Do not check for permissions if user is admin
	if userIsAdmin(user) {
		return nil
	}

	perms := a.getPermissionsForURL(url, method)
	for _, userPerm := range user.Permissions {
		for _, perm := range perms {
			if userPerm == perm {
				return nil
			}
		}
	}

	return ErrPermissionDenied
}

// Returns all permission groups provided URL/Method is allowed for
func (a *authenticator) getPermissionsForURL(url, method string) []string {
	var groups []string
	for groupName, permissions := range a.groupDb {
		for _, permissions := range permissions {
			// Check URL
			if permissions.Url == url {
				// Check allowed methods
				for _, allowed := range permissions.AllowedMethods {
					if allowed == method {
						groups = append(groups, groupName)
					}
				}
			}
		}
	}
	return groups
}

// Checks user admin permission
func userIsAdmin(user *User) bool {
	for _, permission := range user.Permissions {
		if permission == defaultUser {
			return true
		}
	}
	return false
}

var loginFormTmpl = template.Must(template.New("templates").Parse(
	`<!DOCTYPE html>
<html lang="en">
<head>
    <title>Login</title>
</head>
<body>
<div class="form">
    <h2>Login</h2>
    <form method="post" action="authenticate">
        <label for="username">Username</label>
        <input type="text" id="username" name="username"> <br/>
        <label for="password">Password</label>
        <input type="password" id="password" name="password">  <br/>
        <button type="submit">Login</button>
     </form>
</div>
</body>
</html>`))
