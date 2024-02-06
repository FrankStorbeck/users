package users

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Read String Write

func TestNew(t *testing.T) {
	tests := []struct {
		username string
		name     string
		groupIds []int
		err      error
	}{
		{"a@b.c", "A", []int{}, nil},
		{"a@b.c", "A", []int{1, 2}, nil},
		{"a@.c", "A", []int{1, 2}, ErrInvalidUserName},
		{"a@b.c", "A", []int{0}, ErrInvalidGroupId},
	}

	for _, tst := range tests {
		now := time.Now()
		u, err := New(tst.username, tst.name, tst.groupIds)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("New(%q, %q, [%s]) returns an error %s, should be %s",
					tst.username, tst.name, intsString(tst.groupIds), sE1, sE2)
			}
		} else {
			if got := u.UserName(); got != tst.username {
				t.Errorf("UserName() returns to %q, should be %q", got, tst.username)
			}
			if got := u.Name(); got != tst.name {
				t.Errorf("UName() returns to %q, should be %q", got, tst.name)
			}
			if got, want := intsString(u.GroupIds()), intsString(tst.groupIds); got != want {
				t.Errorf("GroupIds() returns [%q], should be [%q]", got, want)
			}
			if got := u.Created(); got.Before(now) {
				t.Errorf("Created() returns %s, should be %s",
					got.Format(time.RFC3339), now.Format(time.RFC3339))
			}
			if got := u.Modified(); got.Before(now) {
				t.Errorf("Modified() returns %s, should be %s",
					got.Format(time.RFC3339), now.Format(time.RFC3339))
			}
		}
	}
}

func TestSet(t *testing.T) {
	u, err := New("a@b.c", "", []int{})
	if err != nil {
		t.Fatalf("New() returns an error: %s", err.Error())
	}

	time.Sleep(time.Second)
	now := time.Now()

	userName := " d@e.f "
	u.SetUserName(userName)
	if got, want := u.UserName(), strings.TrimSpace(userName); got != want {
		t.Errorf("UserName() returns %q, should be %q", got, want)
	}

	name := " D "
	u.SetName(name)
	if got := u.Name(); got != name {
		t.Errorf("UserName() returns %q, should be %q", got, name)
	}

	if got := u.Modified(); got.Before(now) {
		t.Errorf("Modified() returns %s, should be %s",
			got.Format(time.RFC3339), now.Format(time.RFC3339))
	}
}

func TestPassword(t *testing.T) {
	u, err := New("a@b.c", "A", []int{})
	if err != nil {
		t.Fatalf("New() returns an error: %s", err.Error())
	}
	pwd := "a@pNn00tm13s"
	err = u.SetPassword(pwd)
	if err != nil {
		t.Fatalf("SetPassword() returns an error: %s", err.Error())
	}
	err = u.ValidatePassword(pwd)
	if err != nil {
		t.Errorf("ValidatePassword() returns an error: %s, should be nil",
			err.Error())

	}
	err = u.ValidatePassword(pwd + "_")
	if err == nil {
		t.Errorf("ValidatePassword() returns no error, should be %s",
			ErrInvalidPassword)
	} else if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("ValidatePassword() returns error %s, should be %s",
			err.Error(), ErrInvalidPassword)
	}

}

func TestSetGroups(t *testing.T) {
	u, err := New("a@b.c", "A", []int{})
	if err != nil {
		t.Fatalf("New() returns an error: %s", err.Error())
	}

	tests := []struct {
		ids  []int
		want []int
		in   int
		isIn bool
		err  error
	}{
		{[]int{1, 2}, []int{1, 2}, 1, true, nil},
		{[]int{}, []int{}, 1, false, nil},
		{[]int{1, 1}, []int{1}, 2, false, nil},
		{[]int{1, 1, 2, 1, 3}, []int{1, 2, 3}, 0, false, nil},
		{[]int{3, 1, 2, 2, 1, 1}, []int{1, 2, 3}, 2, true, nil},
		{[]int{1, 0}, []int{}, 0, false, ErrInvalidGroupId},
	}

	for _, tst := range tests {
		err := u.SetGroups(tst.ids)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("SetGroups() returns an error %s, should be %s", sE1, sE2)
			}
		} else {
			if got, want := intsString(u.GroupIds()), intsString(tst.want); got != want {
				t.Errorf("Groupids() returns [%s], should be [%s]", got, want)
			}
			if got := u.IsInGroup(tst.in); got != tst.isIn {
				t.Errorf("IsInGroup(%d) returns %t, should be %t",
					tst.in, got, tst.isIn)
			}
		}
	}
}

var (
	usersPath = filepath.Join("testing", ".users.txt")
)

func TestAllString(t *testing.T) {
	s := "a@b.c;*;1;1;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z\n" +
		"b@b.c;*;2;2;B;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z\n" +
		"c@b.c;*;3;1,2;C;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z\n"
	uA, err := parseAll(s)
	if err != nil {
		t.Fatalf("parseAll() returns an error: %s", err.Error())
	}
	sAll, err := uA.String()
	if err != nil {
		t.Fatalf("String() returns an error: %s", err.Error())
	}
	if sAll != s {
		t.Errorf("String() returns\n%q, should be\n%q", sAll, s)
	}
}

func TestDeactivate(t *testing.T) {
	s := `a@b.c;$x$x$xxxxxx;1;1;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z`
	aU, err := parseAll(s)
	if err != nil {
		t.Fatalf("parseAll() returns an error: %s", err.Error())
	}

	tests := []struct {
		selector interface{}
		err      error
	}{
		{"a@b.c", nil},
		{0, ErrNoSuchUser},
	}

	for _, tst := range tests {
		err := aU.Deactivate(tst.selector)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("Deactivate(%v) returns an error: %s, should be: %s",
					tst.selector, sE1, sE2)
			}
		} else {
			sU, _ := aU.Get(tst.selector)
			if sU.hashedPassword != "*" {
				t.Errorf("Deactivate(%v) returns %q; should be \"*\"",
					tst.selector, sU.hashedPassword)
			}
		}
	}
}

func TestPutGetAndUserId(t *testing.T) {
	aU, err := parseAll("")
	if err != nil {
		t.Fatalf("parse() returns an error: %s", err.Error())
	}

	tests := []struct {
		userName string
		name     string
		groupIds []int
		err      error
	}{
		{"A@B.C", "a", []int{1}, nil},
		{"F@B.C", "f", []int{1}, nil},
		{"D@B.C", "d", []int{1, 2}, nil},
		{"E@B.C", "e", []int{2}, nil},
		{"@B.C", "", []int{}, ErrInvalidUserName},
		{"G@B.C", "G", []int{0}, ErrInvalidGroupId},
	}

	uId := 0
	for _, tst := range tests {
		u := User{
			userName: tst.userName,
			name:     tst.name,
			groupIds: tst.groupIds,
		}

		err := aU.Put(&u)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Fatalf("Put(%q) returns an error: %s, should be: %s",
					u.userName, sE1, sE2)
			}
		} else {
			usr, err := aU.Get(u.userName)
			if err != nil {
				t.Fatalf("Get(%q) returns an error: %s, should be nil",
					u.userName, err)
			}

			uId++
			if id := usr.UserId(); id != uId {
				t.Errorf("UserId(%q) returns %d, should be %d", u.userName, id, uId)
			}
		}
	}
}

func TestGetFunc(t *testing.T) {
	s := `a@b.c;*;1;1;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z
b@b.c;*;2;2;B;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z
c@b.c;*;2;1,2;C;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z
`
	l := 3

	aU, err := parseAll(s)
	if err != nil {
		t.Fatalf("parseAll() returns an error: %s", err.Error())
	}

	selectedUsers := aU.GetFunc(func(u User) bool {
		return u.IsInGroup(1)
	})

	if got, want := len(selectedUsers), l-1; got != want {
		t.Errorf("GetFunc() returns %d users, should be %d", got, l)
	} else {
		for _, sU := range selectedUsers {
			if !sU.IsInGroup(1) {
				t.Errorf("selected user %s not in group 1", sU.userName)
			} else {
				u, _ := aU.Get(sU.UserId())
				if !u.IsInGroup(1) {
					t.Errorf("user %s not in group 1", sU.userName)
				}
			}
		}
	}
}

func TestReadAndWrite(t *testing.T) {
	tests := []struct {
		users []User
		key   []byte
	}{
		{
			[]User{
				{
					userName:       "a@b.c",
					hashedPassword: "$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.",
					userId:         2,
					groupIds:       []int{3, 4},
					name:           "A",
					created:        time.Date(2023, time.November, 24, 15, 38, 0, 0, time.UTC),
					modified:       time.Date(2023, time.December, 5, 8, 14, 0, 0, time.UTC),
				},
				{
					userName:       "d@e.f",
					hashedPassword: "*",
					userId:         1,
					groupIds:       []int{},
					name:           "D",
					created:        time.Date(2023, time.November, 24, 14, 25, 0, 0, time.UTC),
					modified:       time.Date(2023, time.December, 5, 8, 14, 0, 0, time.UTC),
				},
			},
			[]byte{}, //("is this a good secret key or not"),
		},
	}

	for _, tst := range tests {
		os.Remove(usersPath)

		aU1 := &AllUsers{}
		for _, usr := range tst.users {
			u, err := New(usr.userName, usr.name, usr.groupIds)
			if err != nil {
				t.Fatalf("New(%q, %q, ...) returns an error: %s",
					usr.userName, usr.name, err.Error())
			}
			err = aU1.Put(&u)
			if err != nil {
				t.Fatalf("Put(...) returns an error: %s", err.Error())
			}
		}

		err := aU1.Write(usersPath, tst.key)
		if err != nil {
			t.Fatalf("Write() returns an error: %s", err.Error())
		}

		aU2, err := Read(usersPath, tst.key)
		if err != nil {
			t.Fatalf("Read() returns an error: %s", err.Error())
		}

		if l1, l2 := len(aU1.usersByEMail), len(aU2.usersByEMail); l1 != l2 {
			t.Fatalf("Read() returns %d users mapped by Id, should be %d", l2, l1)
		}

		if l1, l2 := len(aU1.usersById), len(aU2.usersById); l1 != l2 {
			t.Fatalf("Read() returns %d users mapped by Id, should be %d", l2, l1)
		}

		for k, u2 := range aU2.usersByEMail {
			u1 := aU1.usersByEMail[k]
			if u1S, u2S := u1.String(), u2.String(); u1S != u2S {
				t.Errorf("different usersByEMail[%q]:\n%q and\n%q", k, u1S, u2S)
			}
		}
	}
}

func TestGet(t *testing.T) {
	s := `a@b.c;*;1;1;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z`
	aU, err := parseAll(s)
	if err != nil {
		t.Fatalf("parseAll() returns an error: %s", err.Error())
	}

	tests := []struct {
		selector interface{}
		err      error
	}{
		{1, nil},
		{"a@b.c", nil},
		{0, ErrNoSuchUser},
		{"1", ErrNoSuchUser},
		{2, ErrNoSuchUser},
		{"", ErrNoSuchUser},
	}

	for _, tst := range tests {
		u, err := aU.Get(tst.selector)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("Get(%v) returns an error: %s, should be: %s",
					tst.selector, sE1, sE2)
			}
		} else if sU := u.String(); sU != s {
			t.Errorf("Get(%v) returns\n%s;\nshould be\n%s", tst.selector, sU, s)
		}
	}
}

func TestUserName(t *testing.T) {
	s := "a@b.c;$2a$12$fN63lsa0OxjgxcMpKA6cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1Y;1;3,4;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z\n" +
		"d@e.f;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;2;1;A;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z\n"
	aU, err := parseAll(s)
	if err != nil {
		t.Fatalf("ParseAl() returns an error: %s", err.Error())
	}

	user, err := aU.Get("a@b.c")
	if err != nil {
		t.Fatalf("Get() returns an error: %s", err.Error())
	}

	err = user.SetUserName("e@b.c")
	if err != nil {
		t.Fatalf("SetUserName() returns an error: %s", err.Error())
	}
	aU.Write(usersPath, []byte{})

	_, err = aU.Get("a@b.c")
	if err == nil {
		t.Fatalf("Get() returns no error, should be: %s",
			ErrNoSuchUser.Error())
	} else if !errors.Is(err, ErrNoSuchUser) {
		t.Fatalf("Get() returns wrong error: %s",
			ErrNoSuchUser.Error())
	}

	_, err = aU.Get("e@b.c")
	if err != nil {
		t.Fatalf("Get() returns an error: %s", err.Error())
	}

	_, err = aU.Get("d@e.f")
	if err != nil {
		t.Fatalf("Get() returns an error: %s", err.Error())
	}
}
