// Package auth resolves the HTTP Basic credentials the server requires.
//
// The policy is: authentication is always on. Unless both are given
// explicitly, the username defaults to the OS login account and the password
// is a freshly generated random string, printed once at startup.
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"os/user"
	"strings"
)

// fallbackUsername is used when the OS login account cannot be determined.
const fallbackUsername = "admin"

// passwordBytes is the entropy of a generated password (128 bits).
const passwordBytes = 16

// Credentials are the username/password pair the server checks against.
type Credentials struct {
	Username string
	Password string
	// DefaultedUsername reports whether Username came from the OS login
	// account rather than from --user.
	DefaultedUsername bool
	// GeneratedPassword reports whether Password was randomly generated
	// rather than supplied by the operator. Only a generated password is safe
	// to echo to the console.
	GeneratedPassword bool
}

// Resolve builds the credentials from the (possibly empty) command line
// values, filling in the OS login name and a random password as needed.
func Resolve(username, password string) (Credentials, error) {
	c := Credentials{
		Username: strings.TrimSpace(username),
		Password: password,
	}
	if c.Username == "" {
		c.Username = osUsername()
		c.DefaultedUsername = true
	}
	if c.Password == "" {
		generated, err := GeneratePassword()
		if err != nil {
			return Credentials{}, err
		}
		c.Password = generated
		c.GeneratedPassword = true
	}
	return c, nil
}

// GeneratePassword returns a random, URL-safe password.
func GeneratePassword() (string, error) {
	buf := make([]byte, passwordBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// osUsername returns the login account of the current process owner, stripping
// any Windows "DOMAIN\" prefix, and falls back to the environment and finally
// to a fixed name.
func osUsername() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		name := u.Username
		if i := strings.LastIndexAny(name, `\/`); i >= 0 {
			name = name[i+1:]
		}
		if name != "" {
			return name
		}
	}
	for _, key := range []string{"USER", "USERNAME", "LOGNAME"} {
		if v := strings.TrimSpace(os.Getenv(key)); v != "" {
			return v
		}
	}
	return fallbackUsername
}
