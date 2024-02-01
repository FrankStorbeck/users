package users

import (
	"cmp"
	"net/mail"
	"slices"
)

func isValidEMailAddress(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func mapUser(aU *AllUsers, usr *User) {
	if aU != nil {
		aU.usersByEMail[usr.userName] = usr
		aU.usersById[usr.userId] = usr
		usr.allUsers = aU
	}
}

func selectUser(aU *AllUsers, sOrI interface{}) (*User, bool) {
	var (
		user  *User
		found bool
	)
	switch key := sOrI.(type) {
	case string:
		user, found = aU.usersByEMail[key]
	case int:
		user, found = aU.usersById[key]
	}
	if !found {
		user = &User{}
	}
	return user, found
}

func (aU *AllUsers) sort() []*User {
	users := make([]*User, len(aU.usersByEMail))

	i := 0
	for _, user := range aU.usersByEMail {
		users[i] = user
		i++
	}

	slices.SortFunc(users, func(uA, uB *User) int {
		return cmp.Compare(uA.userId, uB.userId)
	})

	return users
}

func unMapUser(aU *AllUsers, usr *User) {
	if aU != nil {
		delete(aU.usersByEMail, usr.userName)
		delete(aU.usersById, usr.userId)
	}
}
