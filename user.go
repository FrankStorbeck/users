package users

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User holds the data for a user
type User struct {
	created        time.Time // time of creation
	groupIds       []int     // identifiers for the groups, must be positive
	hashedPassword string    // hashed password for the user
	modified       time.Time // last modification time
	name           string    // user's name
	userId         int       // identifier, must be positive
	userName       string    // user name, must be a valid e-mail address
	allUsers       *AllUsers // AllUsers to which this user is a member
}

// Created returns the date and time of creation.
func (u User) Created() time.Time {
	return u.created
}

// GroupIds returns the group id's, a set of unique positive numbers.
func (u User) GroupIds() []int {
	return u.groupIds
}

// IsInGroup returns true if g is is present in the set of group id's.
func (u User) IsInGroup(g int) bool {
	return slices.Contains(u.groupIds, g)
}

// Modified returns the last date and time at which information was
// modified.
func (u User) Modified() time.Time {
	return u.modified
}

// Name returns the name.
func (u User) Name() string {
	return u.name
}

// New returns a new User. userName should be a valid e-mail address,
// otherwise the user name wil become an empty string.
// Only positive and unique group id's will be accepted.
// The user id will have a value of zero, which is an invalid value.
// By putting it in a AllUsers struct, it will be set to a valid one.
// User has a deactivated status.
func New(userName, name string, groupIds []int) (User, error) {
	u := User{
		name:           name,
		hashedPassword: "*",
		created:        time.Now(),
	}
	if !isValidEMailAddress(userName) {
		return u, ErrInvalidUserName
	}
	u.modified = u.created
	err := u.SetGroups(groupIds)

	return u, err
}

// SetGroups sets the group id's. Only positive and unique group id's
// will be accepted.
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

// SetName sets the name.
func (u *User) SetName(name string) {
	u.name = name
	u.modified = time.Now()
}

// SetUserName sets the user name. If the user name is not a valid e-mail
// address ErrInvalidUserName will be returned.
// If user is putted into AllUsers and AllUsers already has a User with than
// user name, ErrUserExists will be returned.
func (u *User) SetUserName(uNameOrId string) error {
	if !isValidEMailAddress(uNameOrId) {
		return ErrInvalidUserName
	}

	if u.allUsers != nil {
		if _, found := selectUser(u.allUsers, uNameOrId); found {
			return ErrUserExists
		}
	}

	unMapUser(u.allUsers, u)
	u.userName = uNameOrId
	mapUser(u.allUsers, u)

	u.modified = time.Now()
	return nil
}

// SetPassword stores a hash of the plain password. If succesfull, it
// return nil.
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
// user id, zero or more group id's separated by comma's, name, time of
// creation and last modification time in RFC3339 format.
func (u User) String() string {
	groups := ""
	for i, grpId := range u.groupIds {
		if i > 0 {
			groups += ","
		}
		groups += strconv.Itoa(grpId)
	}

	return fmt.Sprintf("%s;%s;%d;%s;%s;%s;%s",
		u.userName, u.hashedPassword, u.userId, groups, u.name,
		u.created.Format(time.RFC3339), u.modified.Format(time.RFC3339))
}

// UserId returns the user's identifier.
func (u User) UserId() int {
	return u.userId
}

// UserName returns the user name.
func (u User) UserName() string {
	return u.userName
}

// ValidatePassword validates a password. It returns nil if the password matches.
func (u User) ValidatePassword(plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.hashedPassword), []byte(plainPassword))
}
