package facts_test

import (
	"github.com/coinbase/dexter/facts"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestHostnameContainsFact(t *testing.T) {
	assert := assert.New(t)

	actual, err := os.Hostname()
	assert.Nil(err)

	check, ok := facts.Get("hostname-contains")
	assert.True(ok)

	assert.True(check.Assert([]string{""}))
	assert.True(check.Assert([]string{actual}))
	assert.False(check.Assert([]string{actual + "foobar"}))
}

func TestHostnameIsFact(t *testing.T) {
	assert := assert.New(t)

	actual, err := os.Hostname()
	assert.Nil(err)

	check, ok := facts.Get("hostname-is")
	assert.True(ok)

	assert.False(check.Assert([]string{""}))
	assert.True(check.Assert([]string{actual}))
	assert.False(check.Assert([]string{actual + "foobar"}))
}
