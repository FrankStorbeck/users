package users

import "testing"

func TestCrypt(t *testing.T) {
	key := []byte("is this a good secret key or not")

	s := `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut 
labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco 
laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in 
voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat 
non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
`
	e, err := en(s, key)
	if err != nil {
		t.Fatalf("encode() returns an error: %s", err)
	}

	u, err := de(e, key)
	if err != nil {
		t.Fatalf("decode() returns an error: %s", err)
	}

	if u != s {
		t.Errorf("decoded string not equal to origanal string:\n%s and \n%s", u, s)
	}
}
