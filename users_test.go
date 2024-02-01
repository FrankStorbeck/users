package users

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSetGroups(t *testing.T) {
	user := User{}
	tests := []struct {
		ids  []int
		want []int
		err  error
	}{
		{[]int{1, 2}, []int{1, 2}, nil},
		{[]int{}, []int{}, nil},
		{[]int{1, 1}, []int{1}, nil},
		{[]int{1, 1, 2, 1, 3}, []int{1, 2, 3}, nil},
		{[]int{3, 1, 2, 2, 1, 1}, []int{1, 2, 3}, nil},
		{[]int{1, 0}, []int{}, ErrInvalidGroupId},
	}

	for _, tst := range tests {
		err := user.SetGroups(tst.ids)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("SetGroups() returns an error %s, should be %s", sE1, sE2)
			}
		} else {
			got := user.GroupIds()
			if lg, lw := len(got), len(tst.want); lg != lw {
				t.Errorf("Group() returns %d id's, should be %d", lg, lw)
			} else {
				for i := 0; i < lg; i++ {
					if got[i] != tst.want[i] {
						t.Errorf("Group() returns id %d, should be %d", got[i], tst.want[i])
					}
				}
			}
		}
	}
}

var usersPath = filepath.Join("testing", ".users.txt")

func TestReadAndWrite(t *testing.T) {
	os.Remove(usersPath)

	allUsers1 := &AllUsers{
		usersByEMail: map[string]*User{
			"a@b.c": {
				userId:         2,
				userName:       "a@b.c",
				hashedPassword: "$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.",
				created:        time.Date(2023, time.December, 04, 48, 52, 0, 0, time.UTC),
				modified:       time.Date(2023, time.November, 24, 16, 38, 0, 0, time.UTC),
				name:           "A",
				groupIds:       []int{3, 4},
			},
			"d@e.f": {
				userId:         1,
				userName:       "d@e.f",
				hashedPassword: "$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6",
				modified:       time.Date(2023, time.November, 24, 15, 25, 0, 0, time.UTC),
				created:        time.Date(2023, time.December, 04, 48, 52, 0, 0, time.UTC),
				name:           "D",
				groupIds:       []int{},
			},
		},
	}
	allUsers1.usersById = make(map[int]*User)
	for _, usr := range allUsers1.usersById {
		allUsers1.usersById[usr.userId] = usr
	}

	err := allUsers1.Write(usersPath)
	if err != nil {
		t.Fatalf("Write() returns an error: %s", err.Error())
	}

	allUsers2, err := Read(usersPath)
	if err != nil {
		t.Fatalf("Read() returns an error: %s", err.Error())
	}

	if l1, l2 := len(allUsers1.usersByEMail), len(allUsers2.usersByEMail); l1 != l2 {
		t.Fatalf("Read() returns %d users mapped by EMail, should be %d", l2, l1)
	}

	for k, usr2 := range allUsers2.usersByEMail {
		usr1 := allUsers1.usersByEMail[k]
		if u1S, u2S := usr1.String(), usr2.String(); u1S != u2S {
			t.Errorf("different users for %q: %q and %q", k, u1S, u2S)
		}
	}

	if l1, l2 := len(allUsers1.usersByEMail), len(allUsers2.usersById); l1 != l2 {
		t.Fatalf("Read() returns %d users mapped by Id, should be %d", l2, l1)
	}
}

func TestPut(t *testing.T) {
	os.Remove(usersPath)
	aU, err := Read(usersPath)

	if err != nil {
		t.Fatalf("Read() returns an error: %s", err.Error())
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

	for _, tst := range tests {
		u := User{
			userName: tst.userName,
			name:     tst.name,
			groupIds: tst.groupIds,
		}

		if err := aU.Put(u); err != nil {
			if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
				if len(sE1) > 0 {
					t.Fatalf("Put(%q) returns an error: %s, should be: %s",
						u.String(), sE1, sE2)
				}
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

func TestGetFunc(t *testing.T) {
	s := `a@b.c;*;1;1 ;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z
b@b.c;*;2;2;B;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z
c@b.c;*; 2; 1, 2;C;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z
`
	l := 3

	aU, err := parseAll(s)
	if err != nil {
		t.Fatalf("parseAll() returns an error: %s", err.Error())
	}

	selectedUsers := aU.GetFunc(func(u User) bool {
		return u.IsInGroup(1)
	})

	if lenGot, lWant := len(selectedUsers), l-1; lenGot != lWant {
		t.Errorf("GetFunc() returns %d users, should be %d", lenGot, l)
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
	aU.Write(usersPath)

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
