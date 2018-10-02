package facts_test

import (
	"github.com/coinbase/dexter/engine/helpers"
	"github.com/coinbase/dexter/facts"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunningDockerImageFact(t *testing.T) {
	assert := assert.New(t)

	helpers.StubRunningDockerImages([]string{"ubuntu", "ami"})
	check, ok := facts.Get("running-docker-image")
	assert.True(ok)

	assert.True(check.Assert([]string{"ami"}))
	assert.True(check.Assert([]string{"ami", "no"}))
	assert.True(check.Assert([]string{"ubuntu"}))
	assert.False(check.Assert([]string{"foo"}))
	assert.False(check.Assert([]string{"bunt"}))
}

func TestRunningDockerImageSubstringFact(t *testing.T) {
	assert := assert.New(t)

	helpers.StubRunningDockerImages([]string{"ubuntu", "ami"})
	check, ok := facts.Get("running-docker-image-substring")
	assert.True(ok)

	assert.True(check.Assert([]string{"ami"}))
	assert.True(check.Assert([]string{"ubuntu"}))
	assert.True(check.Assert([]string{""}))
	assert.True(check.Assert([]string{"bunt"}))
	assert.False(check.Assert([]string{"foo"}))
}
