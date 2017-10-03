package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitUser(t *testing.T) {
	namespace, pod, container, user, err := splitUser("foo~bar~baz~nick")
	assert.Nil(t, err)
	assert.Equal(t, "foo", namespace)
	assert.Equal(t, "bar", pod)
	assert.Equal(t, "baz", container)
	assert.Equal(t, "nick", user)
}
