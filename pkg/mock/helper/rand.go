package helper

import "math/rand"

// RandString generate random string from given set
func RandString(s []rune, n int) string {
	if n == 0 {
		return ""
	}
	r := ""
	for i := 0; i < n; i++ {
		r = r + string(s[rand.Intn(len(s))])
	}
	return r
}
