package users

import "testing"

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
		user, err := parse(tst.s)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("parse(%q) returns error %q, should be %q", tst.s, sE1, sE2)
			}
		} else if gS := user.String(); gS != tst.s {
			t.Errorf("parse(%q) returns\n%q,\nshould be\n%q", tst.s, gS, tst.s)
		}
	}
}

func TestParseAll(t *testing.T) {
	tests := []struct {
		s   string
		l   int
		err error
	}{
		{
			"a@b.c;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;1;3,4;A;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z\n" +
				"d@e.f;$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6;2;1;A;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z\n",
			2,
			nil,
		},
		{
			"a@b.c;$2a$12$O82XHvkCrkQzpkr30NNShu81RueblNmjIu6jeZuaGB.d8g7roROI.;0;3,4;D;2023-11-24T15:38:00Z;2023-12-05T08:14:00Z\n" +
				"d@e.f;$2a$12$cKlDQ9UmKhy7XS40fXR8jONaajOX3k1g1YfN63lsa0OxjgxcMpKA6;2;1;A;2023-11-24T16:25:00Z;2023-12-05T08:14:00Z\n",
			2,
			ErrInvalidUserId,
		},
	}

	for _, tst := range tests {
		users, err := parseAll(tst.s)
		if notBothAreNil, sE1, sE2 := testErrs(err, tst.err); notBothAreNil {
			if len(sE1) > 0 {
				t.Errorf("parseAll(%q) returns error %q, should be %q",
					tst.s[:5], sE1, sE2)
			}
		} else if l := len(users.usersByEMail); l != tst.l {
			t.Errorf("parseAll(%q) result has %d users, should be %d", tst.s[:5], l, tst.l)
		}
	}
}
