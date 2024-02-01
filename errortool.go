package users

import "errors"

// tool for testing error results for tests
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
