package users

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// parse creates a User from a string. The string must be formatted as
// the one returned by String().
func parse(s string) (User, error) {
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
			if err != nil || u.userId <= 0 {
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

// parseAll parses all the user data from a string
func parseAll(s string) (*AllUsers, error) {
	aU := &AllUsers{}
	if len(s) == 0 {
		return aU, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return aU, err
		}

		usr, err := parse(scanner.Text())
		if err != nil {
			return aU, err
		}
		aU.mapUser(&usr)
	}

	return aU, nil
}
