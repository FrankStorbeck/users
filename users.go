// package users is a module to manage user data for users that can have access to a server.
// The data are stored in a file like the password file in `*nix`.
//
// The following data are stored: user name, hashed password, user id, zero or more group id's, name, time of creration
// and last time of modification. The user id, and the creation time are immutable. The modification time will change
// when a modification of user name, password or group id's takes place.
//
// To store the data after any change they should be written to a file by calling Write().
package users

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidGroupId  = errors.New("invalid group id")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidUserId   = errors.New("invalid user id")
	ErrInvalidUserName = errors.New("user name is not a valid e-mail address")
	ErrInvalidTime     = errors.New("invalid time")
	ErrMissingData     = errors.New("missing data")
	ErrNoSuchUser      = errors.New("no such user")
	ErrUserExists      = errors.New("user exists")

	mutex sync.Mutex // mutex for reading and writing to file
)

// =========== AllUsers ===========

// AllUsers holds the data of all users for a server.
type AllUsers struct {
	lastId       int              // latest Id used
	usersByEMail map[string]*User // user accounts, the key is the user name
	usersById    map[int]*User    // user accounts, the key is the user id

}

// Deactivate deactivates the user with the provided user name, i.e. calling
// ValidatePassword() will fail afterwards.
func (aU *AllUsers) Deactivate(uNameOrId interface{}) error {
	u, found := selectUser(aU, uNameOrId)
	if !found {
		return ErrNoSuchUser
	}

	u.hashedPassword = "*"
	u.modified = time.Now()
	return nil
}

// Get fetches a user with the provided user name or user id.
func (aU *AllUsers) Get(uNameOrId interface{}) (User, error) {
	var (
		err   error
		u     *User
		found bool
	)
	u, found = selectUser(aU, uNameOrId)

	if !found {
		u, err = &User{}, fmt.Errorf("%w: %s", ErrNoSuchUser, uNameOrId)
	}
	return *u, err
}

// GetFunc returns a slice of users for which f returns true.
func (aU *AllUsers) GetFunc(f func(u User) bool) []User {
	matchingUsers := []User{}

	for _, u := range aU.usersById {
		if f(*u) {
			matchingUsers = append(matchingUsers, *u)
		}
	}

	return matchingUsers
}

// Put puts the user data into u. When an entry for the user is already present
// an error will be returned.
func (aU *AllUsers) Put(u User) error {
	u.userId = 1 // use some dummy value before testing for errors
	if _, err := parse(u.String()); err != nil {
		return err
	}

	if _, found := selectUser(aU, u.userName); found {
		return ErrUserExists
	}

	u.modified = time.Now()
	aU.lastId++
	u.userId = aU.lastId
	mapUser(aU, &u)

	return nil
}

// Read reads the user data from a file located at path. If the
// file doesn't exists, an empty instance of AllUsers will be returned.
func Read(path string) (*AllUsers, error) {
	mutex.Lock()
	defer mutex.Unlock()

	aU := &AllUsers{
		usersByEMail: make(map[string]*User),
		usersById:    make(map[int]*User),
	}

	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
	} else {
		var b []byte
		b, err = io.ReadAll(f)
		if err == nil {
			aU, err = parseAll(string(b))
		}
		f.Close()
	}
	return aU, err
}

// String writes the user data in a string.
func (aU *AllUsers) String() (string, error) {
	sortedUsers := aU.sort()

	var b strings.Builder
	for _, usr := range sortedUsers {
		if _, err := b.WriteString(usr.String() + "\n"); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

// Write stores the user data in a file.
func (aU *AllUsers) Write(path string) error {
	s, err := aU.String()
	if err != nil {
		return err
	}

	mutex.Lock()
	defer mutex.Unlock()

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(s)
	return err
}

// =========== User ===========
