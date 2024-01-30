package users

import (
	"bufio"
	"cmp"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
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
	ErrWeakPassword    = errors.New("password too weak")
)

// =========== User ===========

// User holds the data for a user
type User struct {
	created        time.Time // time of creation
	groupIds       []int     // identifiers for the groups, must be positive
	hashedPassword string    // hashed password for the user
	modified       time.Time // last modification time
	Name           string    // user's name
	userId         int       // identifier, must be positive
	userName       string    // user name, must be a valid e-mail address
	users          *Users    // Users for which this user is a member
}

// Created returns the creation time
func (u User) Created() time.Time {
	return u.created
}

// GroupIds returns the group id's.
func (u *User) GroupIds() []int {
	return u.groupIds
}

// IsInGroup returns true if the user is a member of a group.
func (u User) IsInGroup(g int) bool {
	return slices.Contains(u.groupIds, g)
}

// Modified returns the last date and time at which the user information was
// modified.
func (u User) Modified() time.Time {
	return u.modified
}

// Parse creates a User from a string. The string must be formatted as
// the one returned by String().
func Parse(s string) (User, error) {
	u := User{}
	var err error

	fields := strings.Split(s, ";")
	if l := len(fields); l < 7 {
		return u, fmt.Errorf("%w, less than 7 fields found: %d", ErrMissingData, l)
	}
	fields = fields[:7]

	for i, fld := range fields {
		switch i {
		case 0: // userName, must be a valid e-mail address
			if err = u.SetUserName(fld); err != nil {
				return u, fmt.Errorf("%w (%s)", err, fld)
			}

		case 1: // hashed password
			u.hashedPassword = fld

		case 2: // user id
			u.userId, err = strconv.Atoi(fld)
			if err != nil || u.userId <= 0 {
				return u, fmt.Errorf("%w for user %s: %s", ErrInvalidUserId, u.userName, fld)
			}

		case 3: // group id's
			ids := []int{}

			if len(fld) > 0 {
				gIds := strings.Split(fld, ",")
				for _, gId := range gIds {
					id, err := strconv.Atoi(gId)
					if err != nil {
						return u, fmt.Errorf("%w for user %s: %w",
							ErrInvalidGroupId, u.userName, err)
					}
					ids = append(ids, id)
				}
			}

			if err = u.SetGroups(ids); err != nil {
				return u, fmt.Errorf("cannot set group id's for user %s: %w",
					u.userName, err)
			}

		case 4: // name
			u.Name = fld

		case 5: // creation time
			u.created, err = time.Parse(time.RFC3339, fld)
			if err != nil {
				return u, fmt.Errorf("%w (creation) for user %s: %w",
					ErrInvalidTime, u.userName, err)
			}

		case 6: // modification time
			u.modified, err = time.Parse(time.RFC3339, fld)
			if err != nil {
				return u, fmt.Errorf("%w (modification) for user %s: %w",
					ErrInvalidTime, u.userName, err)
			}
		}
	}

	return u, nil
}

// SetGroups sets the group id's. Id's must be positive.
func (u *User) SetGroups(groupIds []int) error {
	ids := []int{}
	for _, id := range groupIds {
		if id <= 0 {
			return fmt.Errorf("%w: (%d)", ErrInvalidGroupId, id)
		}
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}

	slices.Sort(ids)
	u.groupIds = ids
	u.modified = time.Now()
	return nil
}

// SetUserName sets the user name.
func (u *User) SetUserName(uName string) error {
	if !validEMailAddress(uName) {
		return ErrInvalidUserName
	}

	if u.users != nil {
		if _, found := u.users.users[uName]; found {
			return ErrUserExists
		}
	}

	current := u.userName
	u.userName = uName

	if u.users != nil {
		delete(u.users.users, current)
		u.users.users[uName] = u
	}
	u.modified = time.Now()
	return nil
}

// SetPassword stores a hash of the plain password, but only if the
// score is reached. A reasonable value for score is 30 or more.
// If succesfull, it return nil.
func (u *User) SetPassword(plainPassword string) error {
	b, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 12)
	if err != nil {
		return err
	}

	u.hashedPassword = string(b)
	u.modified = time.Now()
	return nil
}

// String returns a string with the user's information. It holds the
// following fields separated by semi colons: user name, password hash,
// user id, zero or more group id's separated by comma's, name and last
// modification time in RFC3339 format. E.g.
// john.doe@company.com;$2a$12$jRILr9NQVotdxSXeKPVZjfKmkSS5omc1OIhY5e2403uBHc2V30ium;3,4;John Doe:2023-11-24T15:38:00Z;2023-12-01T08:27:00Z
func (u User) String() string {
	groups := ""
	for i, grpId := range u.groupIds {
		if i > 0 {
			groups += ","
		}
		groups += strconv.Itoa(grpId)
	}

	return fmt.Sprintf("%s;%s;%d;%s;%s;%s;%s",
		u.userName, u.hashedPassword, u.userId, groups, u.Name,
		u.created.Format(time.RFC3339), u.modified.Format(time.RFC3339))
}

// UserId returns the user's identifier.
func (u User) UserId() int {
	return u.userId
}

// UserName returns the user name.
func (u *User) UserName() string {
	return u.userName
}

// ValidatePassword validates a password. It returns nil if the password matches.
func (u User) ValidatePassword(plainPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.hashedPassword), []byte(plainPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		err = ErrInvalidPassword
	}
	return err
}

// =========== Users ===========

var mutex sync.Mutex

type Users struct {
	lastId int              // latest Id used
	users  map[string]*User // user accounts, the key is the user name
}

// Deactivate deactivates the user with the provided user name.
func (u *Users) Deactivate(uName string) error {
	user, found := u.users[uName]
	if !found {
		return ErrNoSuchUser
	}

	user.hashedPassword = "*"
	user.modified = time.Now()
	return nil
}

// Get fetches a user with the provided user name.
func (u *Users) Get(uName string) (User, error) {
	var err error
	user, found := u.users[uName]
	if !found {
		return User{}, fmt.Errorf("%w: %s", ErrNoSuchUser, uName)
	}
	return *user, err
}

// GetFunc fetches a slice of users from the users file for which f
// returns true.
func (u *Users) GetFunc(f func(usr User) bool) []User {
	matchingUsers := []User{}

	for _, usr := range u.users {
		if f(*usr) {
			matchingUsers = append(matchingUsers, *usr)
		}
	}

	return matchingUsers
}

// ParseAll parses all the user data from a string
func ParseAll(s string) (*Users, error) {
	u := &Users{users: make(map[string]*User)}

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return u, err
		}

		usr, err := Parse(scanner.Text())
		if err != nil {
			return u, err
		}
		u.users[usr.userName] = &usr
		usr.users = u

		if u.lastId < usr.userId {
			u.lastId = usr.userId
		}
	}

	return u, nil
}

// Put puts the user data into u. When an entry for the user is already present
// an error will be returned.
func (u *Users) Put(user User) error {
	user.userId = u.lastId + 1 // fill in some value before testing for errors
	if _, err := Parse(user.String()); err != nil {
		return err
	}

	if _, found := u.users[user.userName]; found {
		return ErrUserExists
	}

	user.modified = time.Now()
	u.lastId++
	// user.userId = u.lastId
	user.users = u
	u.users[user.userName] = &user

	return nil
}

// Read reads the user data from a file
func Read(path string) (*Users, error) {
	mutex.Lock()
	defer mutex.Unlock()

	u := &Users{users: make(map[string]*User)}

	f, err := os.Open(path)
	if err != nil {
		return u, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if err = scanner.Err(); err != nil {
			return u, err
		}

		usr, err := Parse(scanner.Text())
		if err != nil {
			return u, err
		}
		u.users[usr.userName] = &usr

		if u.lastId < usr.userId {
			u.lastId = usr.userId
		}
	}

	return u, nil
}

// String writes the user data in a string.
func (u *Users) String() (string, error) {
	users := u.sort()

	var b strings.Builder
	for _, usr := range users {
		if _, err := b.WriteString(usr.String() + "\n"); err != nil {
			return "", err
		}
	}
	return b.String(), nil
}

// Update updates the user data.
func (u *Users) Update(user User) error {
	if _, found := u.users[user.userName]; !found {
		return ErrNoSuchUser
	}
	user.modified = time.Now()
	user.users = u

	u.users[user.userName] = &user
	return nil
}

// Write stores the user data in a file.
func (u *Users) Write(path string) error {
	mutex.Lock()
	defer mutex.Unlock()

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	users := u.sort()

	for _, usr := range users {
		if _, err := f.WriteString(usr.String() + "\n"); err != nil {
			return err
		}
	}

	return nil
}

// =========== tools ===========

func (u *Users) sort() []*User {
	users := make([]*User, len(u.users))

	i := 0
	for _, user := range u.users {
		users[i] = user
		i++
	}

	slices.SortFunc(users, func(uA, uB *User) int {
		return cmp.Compare(uA.userId, uB.userId)
	})

	return users
}

func validEMailAddress(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}
