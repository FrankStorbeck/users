package users

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User holds the data for a user
type User struct {
	allUsers       *AllUsers // AllUsers containing this user
	created        time.Time // time of creation
	groupIds       []int     // identifiers for the groups, must be positive
	hashedPassword string    // hashed password for the user
	modified       time.Time // last modification time
	name           string    // user's name
	userId         int       // identifier, must be positive
	userName       string    // user name, must be a valid e-mail address
}

// Deactivate deactivates the user.
func (u *User) Deactivate() {
	if u.hashedPassword[:1] == "*" {
		return
	}

	u.hashedPassword = "*" + u.hashedPassword
	u.modified = time.Now()
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

// New returns a new User. userName should be a valid e-mail address.
// Only positive and unique group id's will be accepted.
// The user id will have a value of zero. After putting it in a AllUsers
// struct, it will be set to a unique value.
// User has a deactivated status.
func New(userName, name string, groupIds []int) (User, error) {
	u := User{
		userName:       strings.TrimSpace(userName),
		name:           name,
		hashedPassword: "*",
		created:        time.Now(),
	}
	u.modified = u.created
	if !isValidEMailAddress(u.userName) {
		return u, ErrInvalidUserName
	}

	return u, u.SetGroups(groupIds)
}

// Parse creates single User instance by parsing a string. The string must be formatted
// accordingly to the one as returned by String().
func Parse(s string) (User, error) {
	u := User{}
	var err error

	fields := strings.Split(s, ";")
	if l := len(fields); l < 7 {
		return u, fmt.Errorf("%w, less than 7 fields found: %d", ErrMissingData, l)
	}
	fields = fields[:7]

	for i, fld := range fields {
		fld = strings.TrimSpace(fld)

		switch i {
		case 0: // userName, must be a valid e-mail address
			if err = u.SetUserName(fld); err != nil {
				return u, fmt.Errorf("%w (%s)", err, fld)
			}

		case 1: // hashed password
			u.hashedPassword = fld

		case 2: // user id
			u.userId, err = strconv.Atoi(fld)
			if err != nil || u.userId < 0 {
				return u, fmt.Errorf("%w for user %s: %s", ErrInvalidUserId, u.userName, fld)
			}

		case 3: // group id's
			ids := []int{}

			if len(fld) > 0 {
				gIds := strings.Split(fld, ",")
				for _, gId := range gIds {
					id, err := strconv.Atoi(strings.TrimSpace(gId))
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
			u.name = fld

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

// Reactivate reactivates the user with the provided user name or user id.
// can be validated again.
func (u *User) Reactivate() {
	if len(u.hashedPassword) == 0 || u.hashedPassword[:1] != "*" {
		return
	}

	u.hashedPassword = u.hashedPassword[1:]
	u.modified = time.Now()
}

// SetGroups sets the group id's. Only non negative and unique group id's
// will be accepted.
func (u *User) SetGroups(groupIds []int) error {
	ids := []int{}
	for _, id := range groupIds {
		if id < 0 {
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

// SetPassword stores a hash of the plain password. If succesfull, it
// returns nil.
func (u *User) SetPassword(plainPassword string) error {
	b, err := bcrypt.GenerateFromPassword([]byte(plainPassword), 12)
	if err != nil {
		return err
	}

	u.hashedPassword = string(b)
	u.modified = time.Now()
	return nil
}

// SetUserName sets the user name. If the user name is not a valid e-mail
// address ErrInvalidUserName will be returned.
// If user is putted into AllUsers and AllUsers already has a User with than
// user name, ErrUserExists will be returned.
func (u *User) SetUserName(uName string) error {
	uName = strings.TrimSpace(uName)
	if !isValidEMailAddress(uName) {
		return ErrInvalidUserName
	}

	if u.allUsers != nil {
		if _, found := selectUser(u.allUsers, uName); found {
			return ErrUserExists
		}

		u.allUsers.unMapUser(u)
		u.userName = uName
		u.allUsers.mapUser(u)
	} else {
		u.userName = uName
	}

	u.modified = time.Now()
	return nil
}

// String returns a string with the user's information. It holds the
// following fields separated by semi colons: user name, password hash,
// user id, zero or more group id's separated by comma's, name, time of
// creation and last modification time in RFC3339 format.
func (u User) String() string {
	return fmt.Sprintf("%s;%s;%d;%s;%s;%s;%s",
		u.userName, u.hashedPassword, u.userId, intsString(u.groupIds), u.name,
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
	err := bcrypt.CompareHashAndPassword([]byte(u.hashedPassword), []byte(plainPassword))
	if err != nil {
		err = fmt.Errorf("%w: %w", ErrInvalidPassword, err)
	}
	return err
}
