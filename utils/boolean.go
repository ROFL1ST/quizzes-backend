package utils

import "strings"

// IsBooleanMatch compares two strings to see if they represent the same boolean truth value.
// It normalizes True-like ("true", "benar", "yes", "1") and False-like ("false", "salah", "no", "0") values.
func IsBooleanMatch(ans1, ans2 string) bool {
	a1 := strings.ToLower(strings.TrimSpace(ans1))
	a2 := strings.ToLower(strings.TrimSpace(ans2))

	isTrue := func(s string) bool {
		return s == "true" || s == "benar" || s == "yes" || s == "1"
	}
	isFalse := func(s string) bool {
		return s == "false" || s == "salah" || s == "no" || s == "0"
	}

	if isTrue(a1) && isTrue(a2) {
		return true
	}
	if isFalse(a1) && isFalse(a2) {
		return true
	}
	return a1 == a2
}
