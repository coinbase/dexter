package facts_test

import (
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/coinbase/dexter/facts"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProjectNameContainsFact(t *testing.T) {
	assert := assert.New(t)

	helpers.StubProjectName("foo/bar")
	actual := helpers.ProjectName()
	assert.Equal("foo/bar", actual)

	check, ok := facts.Get("project-name-contains")
	assert.True(ok)

	assert.True(check.Assert([]string{""}))
	assert.True(check.Assert([]string{"foo", "no"}))
	assert.True(check.Assert([]string{actual}))
	assert.False(check.Assert([]string{actual + "extra"}))
}

func TestProjectNameIsFact(t *testing.T) {
	assert := assert.New(t)

	helpers.StubProjectName("foo/bar")
	actual := helpers.ProjectName()
	assert.Equal("foo/bar", actual)

	check, ok := facts.Get("project-name-is")
	assert.True(ok)

	assert.False(check.Assert([]string{""}))
	assert.False(check.Assert([]string{"foo"}))
	assert.True(check.Assert([]string{actual}))
	assert.False(check.Assert([]string{actual + "extra"}))
}
