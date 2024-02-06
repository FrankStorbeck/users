package users

import (
	"cmp"
	"net/mail"
	"slices"
	"strconv"
)

func intsString(ints []int) (s string) {
	sep := ""
	for _, i := range ints {
		s += sep + strconv.Itoa(i)
		sep = ","
	}
	return
}

func isValidEMailAddress(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func (aU *AllUsers) mapUser(u *User) {
	if aU.usersByEMail == nil || aU.usersById == nil {
		aU.usersByEMail = make(map[string]*User)
		aU.usersById = make(map[int]*User)
	}

	if u.userId == 0 {
		aU.lastId++
		u.userId = aU.lastId
	} else if u.userId > aU.lastId {
		aU.lastId = u.userId
	}

	aU.usersByEMail[u.userName] = u
	aU.usersById[u.userId] = u
	u.allUsers = aU

}

func selectUser(aU *AllUsers, sOrI interface{}) (*User, bool) {
	var (
		u     *User
		found bool
	)
	switch key := sOrI.(type) {
	case string:
		u, found = aU.usersByEMail[key]
	case int:
		u, found = aU.usersById[key]
	}
	if !found {
		u = &User{}
	}
	return u, found
}

func (aU *AllUsers) sort() []*User {
	users := make([]*User, len(aU.usersById))

	i := 0
	for _, user := range aU.usersById {
		users[i] = user
		i++
	}

	slices.SortFunc(users, func(uA, uB *User) int {
		return cmp.Compare(uA.userId, uB.userId)
	})

	return users
}

func (aU *AllUsers) unMapUser(u *User) {
	if aU.usersByEMail != nil && aU.usersById != nil {
		delete(aU.usersByEMail, u.userName)
		delete(aU.usersById, u.userId)
	}
}
