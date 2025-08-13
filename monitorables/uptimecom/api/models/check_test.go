package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck_HasTag(t *testing.T) {
	check := &Check{Tags: []string{"tag1", "tag2"}}
	assert.True(t, check.MatchOneTag([]string{"tag1"}))
	assert.True(t, check.MatchOneTag([]string{"tag1", "tag2"}))
	assert.True(t, check.MatchOneTag([]string{"tag1", "tag3"}))
	assert.False(t, check.MatchOneTag([]string{"tag3"}))
	assert.False(t, check.MatchOneTag([]string{"tag3", "tag4"}))
}
