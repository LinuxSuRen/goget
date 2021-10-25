package common

import "strings"

// IsValid check if this is a valid uri
func IsValid(uri string) bool {
	return strings.HasPrefix(uri, "/github.com/") ||
		strings.HasPrefix(uri, "/gitee.com/")
}
