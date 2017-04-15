package main

import (
	"fmt"
	"strings"
)

// Used for mashalling a username.
//  * Namespace
//  * Pod
//  * Container
//  * User
func splitUser(u string) (string, string, string, string, error) {
	sl := strings.Split(u, separator)

	if len(sl) == 4 {
		return sl[0], sl[1], sl[2], sl[3], nil
	}

	return "", "", "", "", fmt.Errorf("Failed to marshal string: %s", u)
}
