package users

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGroup(t *testing.T) {
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

func TestParse(t *testing.T) {
	tests := []struct {
		s   string
		err error
	}{
		{
			"a@b.c;$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6;1;3,4;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z",
			nil,
		},
		{
			"d@e.f;$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6;2;1;A;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z",
			nil,
		},
		{
			"a@.c;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;3;3;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z",
			ErrInvalidUserName,
		},
		{
			"a@b.c;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;-1;3;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z",
			ErrInvalidUserId,
		},
		{
			"a@b.c;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;o;3;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z",
			ErrInvalidUserId,
		},
		{
			"a@b.c;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;1;-3,9;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z",
			ErrInvalidGroupId,
		},
		{
			"d@e.f;$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6;2;1;A;2023-11-24xx16:25:00Z;2023-12-05T08:14:00Z",
			ErrInvalidTime,
		},
	}

	for _, tst := range tests {
		user, err := Parse(tst.s)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("Parse(%q) returns error %q, should be %q", tst.s, sE1, sE2)
			}
		} else if gS := user.String(); gS != tst.s {
			t.Errorf("Parse(%q) returns\n%q,\nshould be\n%q", tst.s, gS, tst.s)
		}
	}
}

func TestParseAllAndString(t *testing.T) {
	s := "a@b.c;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;1;3,4;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z\n" +
		"d@e.f;$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6;2;1;A;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z\n"
	users, err := ParseAll(s)
	if err != nil {
		t.Fatalf("ParseAl() returns an error: %s", err.Error())
	}
	got, err := users.String()
	if err != nil {
		t.Fatalf("String() returns an error: %s", err.Error())
	}
	if got != s {
		t.Errorf("String() returns\n%s\nshould be\n%s", got, s)
	}
}

func TestParseGroup(t *testing.T) {
	s := "d@e.f;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;2;3,1,2,1,3,2;A;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z"
	want := "d@e.f;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;2;1,2,3;A;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z"

	user, err := Parse(s)
	if err != nil {
		t.Errorf("Parse(%q) returns an error: %s", s, err)
	} else if uS := user.String(); uS != want {
		t.Errorf("Parse(%q) returns %q, should be %q", s, uS, want)
	}
}

var usersPath = filepath.Join("testing", ".users.txt")

func TestReadAndWrite(t *testing.T) {
	os.Remove(usersPath)

	users1 := &Users{
		users: map[string]*User{
			"a@b.c": {
				userId:         2,
				userName:       "a@b.c",
				hashedPassword: "$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.",
				modified:       time.Date(2023, time.November, 24, 16, 38, 0, 0, time.UTC),
				Name:           "A",
				groupIds:       []int{3, 4},
			},
			"d@e.f": {
				userId:         1,
				userName:       "d@e.f",
				hashedPassword: "$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6",
				modified:       time.Date(2023, time.November, 24, 15, 25, 0, 0, time.UTC),
				Name:           "D",
				groupIds:       []int{},
			},
		},
	}

	err := users1.Write(usersPath)
	if err != nil {
		t.Fatalf("Write() returns an error: %s", err.Error())
	}

	users2, err := Read(usersPath)
	if err != nil {
		t.Fatalf("Read() returns an error: %s", err.Error())
	}

	if l1, l2 := len(users1.users), len(users2.users); l1 != l2 {
		t.Fatalf("Read() returns %d users, should be %d", l2, l1)
	}

	for k, usr1 := range users1.users {
		usr2 := users2.users[k]
		if u1S, u2S := usr1.String(), usr2.String(); u1S != u2S {
			t.Errorf("different users for %q: %q and %q", k, u1S, u2S)
		}
	}
}

func TestPutAndGetFunc(t *testing.T) {
	os.Remove(usersPath)
	users, err := Read(usersPath)

	if err != nil {
		t.Fatalf("Read() returns an error: %s", err.Error())
	}

	newUsers := []User{
		{
			userName: "A@B.C",
			Name:     "a",
			groupIds: []int{1},
		},
		{
			userName: "D@B.C",
			Name:     "d",
			groupIds: []int{1, 2},
		},
		{
			userName: "E@B.C",
			Name:     "e",
			groupIds: []int{2},
		},
		{
			userName: "F@B.C",
			Name:     "f",
			groupIds: []int{1},
		},
	}

	now := time.Now()

	for _, user := range newUsers {
		if err := users.Put(user); err != nil {
			t.Fatalf("Put() returns an error: %s", err.Error())
		}
	}

	selectedUsers := users.GetFunc(func(u User) bool {
		return u.IsInGroup(1)
	})

	if notBothAreNil, sE1, sE2 := testErrs(err, nil); notBothAreNil {
		if len(sE1) > 0 {
			t.Errorf("GetFunc() returns error %s, should be %s", sE1, sE2)
		}
	} else if lenGot, lenWant := len(selectedUsers), len(newUsers)-1; lenGot != lenWant {
		t.Errorf("GetFunc() returns %d users, should be %d", lenGot, lenWant)
	} else {
		for i, user := range newUsers {
			if user.IsInGroup(1) {
				id := i + 1
				j := 0

				for j < lenGot {
					if selectedUsers[j].UserId() == id {
						break
					}
					j++
				}

				if j >= lenGot {
					t.Errorf("GetFunc() should contain %q but didn't", user.String())
				} else if selectedUsers[j].userName != user.userName {
					t.Errorf("GetFunc() returns\n%q,\nshould be\n%q", selectedUsers[j].String(), user.String())
				} else {
					if m := selectedUsers[j].Modified(); now.After(m) {
						t.Errorf("Modified() returns %s, should be after %s",
							m.Format(time.RFC3339), now.Format(time.RFC3339))
					}

				}
			}
		}
	}
}

func TestPutGetReplaceAndDeactivate(t *testing.T) {
	os.Remove(usersPath)
	users, err := Read(usersPath)
	if err != nil {
		t.Fatalf("Read() returns an error: %s", err.Error())
	}

	tests := []struct {
		funcName string
		uS       string
		uName    string
		err      error
	}{
		{"Put", "g@h.i;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;4;3;G;2023-11-24T18:22:01Z;2023-12-05T08:14:00Z", "", nil},
		{"Get", "g@h.i;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;4;3;G;2023-11-24T18:22:01Z;2023-12-05T08:14:00Z", "g@h.i", nil},
		{"Deactivate", "", "g@h.i", nil},
		{"Get", "", "g@h.i", nil},
		{"Put", "j@k.l;$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6;5;1;J;2023-11-24T18:22:01Z;2023-12-05T08:14:00Z", "", nil},
		{"Put", "j@k.l;$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6;5;1;J;2023-11-24T18:22:01Z;2023-12-05T08:14:00Z", "", ErrUserExists},
		{"Deactivate", "", "m@n.p", ErrNoSuchUser},
		{"Update", "j@k.l;$2a$12$fN63lsa0OxjgxcMpKA6cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1Y;5;1;J;2023-11-24T18:22:01Z;2023-12-05T08:14:00Z", "", nil},
		{"Update", "q@r.s;$2a$12$fN63lsa0OxjgxcMpKA6cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1Y;5;1;J;2023-11-24T18:22:01Z;2023-12-05T08:14:00Z", "", ErrNoSuchUser},
	}

	for _, tst := range tests {
		var gErr error

		switch tst.funcName {
		case "Put":
			user, err := Parse(tst.uS)
			if err != nil {
				t.Fatalf("Parse(%q) returns an error: %q, should be nil", tst.uS, err.Error())
			}
			gErr = users.Put(user)
		case "Get":
			_, gErr = users.Get(tst.uName)
		case "Update":
			user, err := Parse(tst.uS)
			if err != nil {
				t.Fatalf("Parse(%q) returns an error: %q, should be nil", tst.uS, err.Error())
			}
			gErr = users.Update(user)
		case "Deactivate":
			gErr = users.Deactivate(tst.uName)
		}

		if notBothAreNil, sE1, sE2 := testErrs(gErr, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("%s() returns %q, should be %q", tst.funcName, sE1, sE2)
			}
		}
	}
}

func TestUserName(t *testing.T) {
	s := "a@b.c;$2a$12$fN63lsa0OxjgxcMpKA6cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1Y;1;3,4;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z\n" +
		"d@e.f;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;2;1;A;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z\n"
	users, err := ParseAll(s)
	if err != nil {
		t.Fatalf("ParseAl() returns an error: %s", err.Error())
	}

	user, err := users.Get("a@b.c")
	if err != nil {
		t.Fatalf("Get() returns an error: %s", err.Error())
	}

	err = user.SetUserName("e@b.c")
	if err != nil {
		t.Fatalf("SetUserName() returns an error: %s", err.Error())
	}

	_, err = users.Get("a@b.c")
	if err == nil {
		t.Fatalf("SetUserName() returns no error, should be: %s",
			ErrNoSuchUser.Error())
	} else if !errors.Is(err, ErrNoSuchUser) {
		t.Fatalf("SetUserName() returns wrong error: %s",
			ErrNoSuchUser.Error())
	}

	_, err = users.Get("e@b.c")
	if err != nil {
		t.Fatalf("Get() returns an error: %s", err.Error())
	}

	_, err = users.Get("d@e.f")
	if err != nil {
		t.Fatalf("Get() returns an error: %s", err.Error())
	}
}

// =========== tools ===========

// tool for testing error results
func testErrs(e1, e2 error) (notBothAreNil bool, sE1 string, sE2 string) {
	if e1 == nil && e2 == nil {
		return
	}
	notBothAreNil = true
	if e1 == nil {
		sE1 = "nil"
		sE2 = e2.Error()
	} else if e2 == nil {
		sE1 = e1.Error()
		sE2 = "nil"
	} else if !errors.Is(e1, e2) {
		sE1 = e1.Error()
		sE2 = e2.Error()
	}
	return
}
