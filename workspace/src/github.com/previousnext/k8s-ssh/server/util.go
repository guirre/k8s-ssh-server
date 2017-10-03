package main

import (
	"fmt"
	"strings"
)

const separator = "~"

// Used for mashalling a ssh username into
//  * Namespace
//  * Pod
//  * Container
//  * User
func splitUser(user string) (string, string, string, string, error) {
	sl := strings.Split(user, separator)

	if len(sl) == 4 {
		return sl[0], sl[1], sl[2], sl[3], nil
	}

	return "", "", "", "", fmt.Errorf("failed to marshal string: %s", user)
}

// Helper function to determine if the command = shell.
func isShell(cmd []string) bool {
	if len(cmd) == 0 {
		return true
	}

	return false
}

// Helper function to determine if user is rsyncing.
func isRsync(cmd []string) bool {
	if len(cmd) > 0 && cmd[0] == "rsync" {
		return true
	}

	return false
}
