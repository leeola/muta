//
// # General utility funcs
//
package utils

// Check if a string slice contains the given string
func ContainsString(ss []string, s string) bool {
	for _, e := range ss {
		if s == e {
			return true
		}
	}
	return false
}
